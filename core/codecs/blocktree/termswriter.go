package blocktree

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"slices"
	"sync/atomic"

	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/core/codecs"
	"github.com/geange/lucene-go/core/codecs/types"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	coreIndex "github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	"github.com/geange/lucene-go/core/util/array"
	"github.com/geange/lucene-go/core/util/compress"
	"github.com/geange/lucene-go/core/util/fst"
	"github.com/geange/lucene-go/core/util/packed"
)

var _ index.FieldsConsumer = &TermsWriter{}

type TermsWriter struct {
	metaOut         store.IndexOutput
	termsOut        store.IndexOutput
	indexOut        store.IndexOutput
	maxDoc          int
	minItemsInBlock int
	maxItemsInBlock int
	postingsWriter  types.PostingsWriter
	fieldInfos      coreIndex.FieldInfos
	fields          []*store.BufferDataOutput

	closed *atomic.Bool

	scratchBytes *store.RAMOutputStream
}

// NewTermsWriter
// Block-based terms index and dictionary writer.
// Writes terms dict and index, block-encoding (column stride) each term's metadata for each set of terms between two index terms.
// Files:
// .tim: Term Dictionary
// .tip: Term Index
//
// Term Dictionary
// The .tim file contains the list of terms in each field along with per-term statistics (such as docfreq) and per-term metadata (typically pointers to the postings list for that term in the inverted index).
// The .tim is arranged in blocks: with blocks containing a variable number of entries (by default 25-48), where each entry is either a term or a reference to a sub-block.
// NOTE: The term dictionary can plug into different postings implementations: the postings writer/reader are actually responsible for encoding and decoding the Postings Metadata and Term Metadata sections.
// TermsDict (.tim) --> Header, PostingsHeader, NodeBlockNumBlocks, FieldSummary, DirOffset, Footer
// NodeBlock --> (OuterNode | InnerNode)
// OuterNode --> EntryCount, SuffixLength, ByteSuffixLength, StatsLength, < TermStats >EntryCount, MetaLength, <TermMetadata>EntryCount
// InnerNode --> EntryCount, SuffixLength[,Sub?], ByteSuffixLength, StatsLength, < TermStats ? >EntryCount, MetaLength, <TermMetadata ? >EntryCount
// TermStats --> DocFreq, TotalTermFreq
// FieldSummary --> NumFields, <FieldNumber, NumTerms, RootCodeLength, ByteRootCodeLength, SumTotalTermFreq?, SumDocFreq, DocCount, LongsSize, MinTerm, MaxTerm>NumFields
// Header --> CodecHeader
// DirOffset --> Uint64
// MinTerm,MaxTerm --> VInt length followed by the byte[]
// EntryCount,SuffixLength,StatsLength,DocFreq,MetaLength,NumFields, FieldNumber,RootCodeLength,DocCount,LongsSize --> VInt
// TotalTermFreq,NumTerms,SumTotalTermFreq,SumDocFreq --> VLong
// Footer --> CodecFooter
// Notes:
// Header is a CodecHeader storing the version information for the BlockTree implementation.
// DirOffset is a pointer to the FieldSummary section.
// DocFreq is the count of documents which contain the term.
// TotalTermFreq is the total number of occurrences of the term. This is encoded as the difference between the total number of occurrences and the DocFreq.
// FieldNumber is the fields number from FieldInfos. (.fnm)
// NumTerms is the number of unique terms for the field.
// RootCode points to the root block for the field.
// SumDocFreq is the total number of postings, the number of term-document pairs across the entire field.
// DocCount is the number of documents that have at least one posting for this field.
// LongsSize records how many long values the postings writer/reader record per term (e.g., to hold freq/prox/doc file offsets).
// MinTerm, MaxTerm are the lowest and highest term in this field.
// PostingsHeader and TermMetadata are plugged into by the specific postings implementation: these contain arbitrary per-file data (such as parameters or versioning information) and per-term data (such as pointers to inverted files).
// For inner nodes of the tree, every entry will steal one bit to mark whether it points to child nodes(sub-block). If so, the corresponding TermStats and TermMetaData are omitted
//
// Term Index
// The .tip file contains an index into the term dictionary, so that it can be accessed randomly. The index is also used to determine when a given term cannot exist on disk (in the .tim file), saving a disk seek.
// TermsIndex (.tip) --> Header, FSTIndexNumFields <IndexStartFP>NumFields, DirOffset, Footer
// Header --> CodecHeader
// DirOffset --> Uint64
// IndexStartFP --> VLong
//
// FSTIndex --> FST<byte[]>
// Footer --> CodecFooter
// Notes:
// The .tip file contains a separate FST for each field. The FST maps a term prefix to the on-disk block that holds all terms starting with that prefix. Each field's IndexStartFP points to its FST.
// DirOffset is a pointer to the start of the IndexStartFPs for all fields
// It's possible that an on-disk block would contain too many terms (more than the allowed maximum (default: 48)). When this happens, the block is sub-divided into new blocks (called "floor blocks"), and then the output in the FST for the block's prefix encodes the leading byte of each sub-block, and its file pointer.
func NewTermsWriter(ctx context.Context, state *index.SegmentWriteState, postingsWriter types.PostingsWriter,
	minItemsInBlock, maxItemsInBlock int) (*TermsWriter, error) {

	maxDoc, err := state.SegmentInfo.MaxDoc()
	if err != nil {
		return nil, err
	}

	this := &TermsWriter{
		minItemsInBlock: minItemsInBlock,
		maxItemsInBlock: maxItemsInBlock,
		maxDoc:          maxDoc,
		fieldInfos:      state.FieldInfos,
		postingsWriter:  postingsWriter,
		closed:          new(atomic.Bool),
	}

	termsName := store.SegmentFileName(state.SegmentInfo.Name(), state.SegmentSuffix, TERMS_EXTENSION)
	termsOut, err := state.Directory.CreateOutput(ctx, termsName)
	if err != nil {
		return nil, err
	}
	this.termsOut = termsOut

	if err := codecs.WriteIndexHeader(ctx, termsOut, TERMS_CODEC_NAME, VERSION_CURRENT,
		state.SegmentInfo.GetId(), state.SegmentSuffix); err != nil {
		return nil, err
	}

	indexName := store.SegmentFileName(state.SegmentInfo.Name(), state.SegmentSuffix, TERMS_INDEX_EXTENSION)
	indexOut, err := state.Directory.CreateOutput(ctx, indexName)
	if err != nil {
		return nil, err
	}
	this.indexOut = indexOut

	if err := codecs.WriteIndexHeader(ctx, indexOut, TERMS_INDEX_CODEC_NAME, VERSION_CURRENT,
		state.SegmentInfo.GetId(), state.SegmentSuffix); err != nil {
		return nil, err
	}

	metaName := store.SegmentFileName(state.SegmentInfo.Name(), state.SegmentSuffix, TERMS_META_EXTENSION)
	metaOut, err := state.Directory.CreateOutput(ctx, metaName)
	if err != nil {
		return nil, err
	}
	this.metaOut = metaOut

	if err := codecs.WriteIndexHeader(ctx, metaOut, TERMS_META_CODEC_NAME, VERSION_CURRENT,
		state.SegmentInfo.GetId(), state.SegmentSuffix); err != nil {
		return nil, err
	}

	if err := postingsWriter.Init(ctx, metaOut, state); err != nil { // have consumer write its format/header
		return nil, err
	}

	return this, nil
}

func (t *TermsWriter) Close() error {
	if !t.closed.CompareAndSwap(false, true) {
		return nil
	}

	ctx := context.Background()

	if err := t.metaOut.WriteUvarint(ctx, uint64(len(t.fields))); err != nil {
		return err
	}
	for _, fieldMeta := range t.fields {
		fieldMeta.CopyTo(t.metaOut)
	}
	if err := codecs.WriteFooter(ctx, t.indexOut); err != nil {
		return err
	}
	if err := t.metaOut.WriteUvarint(ctx, uint64(t.indexOut.GetFilePointer())); err != nil {
		return err
	}
	if err := codecs.WriteFooter(ctx, t.termsOut); err != nil {
		return err
	}
	if err := t.metaOut.WriteUvarint(ctx, uint64(t.termsOut.GetFilePointer())); err != nil {
		return err
	}
	if err := codecs.WriteFooter(ctx, t.metaOut); err != nil {
		return err
	}

	return util.Close(t.metaOut, t.termsOut, t.indexOut, t.postingsWriter)
}

func (t *TermsWriter) Write(ctx context.Context, fields coreIndex.Fields, norms coreIndex.NormsProducer) error {
	//var lastField string
	for field := range fields.Iterator() {
		//assert lastField == null || lastField.compareTo(field) < 0;
		//lastField = field

		//if (DEBUG) System.out.println("\nBTTW.write seg=" + segment + " field=" + field);
		terms, _ := fields.Terms(field)
		if terms == nil {
			continue
		}

		termsEnum, err := terms.Iterator()
		if err != nil {
			return err
		}
		tw := t.newTermsWriter(t.fieldInfos.FieldInfo(field))
		for {
			term, err := termsEnum.Next(ctx)
			//if (DEBUG) System.out.println("BTTW: next term " + term);

			if err != nil {
				break
			}

			//if (DEBUG) System.out.println("write field=" + fieldInfo.name + " term=" + brToString(term));
			if err := tw.Write(ctx, term, termsEnum, norms); err != nil {
				return err
			}
		}

		if err := tw.Finish(ctx); err != nil {
			return err
		}

		//if (DEBUG) System.out.println("\nBTTW.write done seg=" + segment + " field=" + field);
	}
	return nil
}

type pendingEntry interface {
	IsTerm() bool
}

var _ pendingEntry = &pendingTerm{}

type pendingTerm struct {
	termBytes []byte
	state     types.BlockTermState
	isTerm    bool
}

func newPendingTerm(term []byte, state types.BlockTermState) *pendingTerm {
	return &pendingTerm{
		isTerm:    true,
		termBytes: slices.Clone(term),
		state:     state,
	}
}

func (p *pendingTerm) IsTerm() bool {
	return p.isTerm
}

var _ pendingEntry = &pendingBlock{}

type pendingBlock struct {
	prefix        []byte
	fp            int64
	index         *fst.FST[[]byte]
	subIndices    []*fst.FST[[]byte]
	hasTerms      bool
	isFloor       bool
	floorLeadByte int
	isTerm        bool
}

func newPendingBlock(prefix []byte, fp int64, hasTerms bool, isFloor bool,
	floorLeadByte int, subIndices []*fst.FST[[]byte]) *pendingBlock {
	return &pendingBlock{
		isTerm:        false,
		prefix:        prefix,
		fp:            fp,
		hasTerms:      hasTerms,
		isFloor:       isFloor,
		floorLeadByte: floorLeadByte,
		subIndices:    subIndices,
	}
}

func (p *pendingBlock) IsTerm() bool {
	return p.isTerm
}

func (p *pendingBlock) compileIndex(ctx context.Context, blocks []*pendingBlock, scratchBytes *store.RAMOutputStream) error {
	// TODO: try writing the leading vLong in MSB order
	// (opposite of what Lucene does today), for better
	// outputs sharing in the FST
	if err := scratchBytes.WriteUvarint(ctx, encodeOutput(p.fp, p.hasTerms, p.isFloor)); err != nil {
		return err
	}
	if p.isFloor {
		if err := scratchBytes.WriteUvarint(ctx, uint64(len(blocks)-1)); err != nil {
			return err
		}
		for i := 1; i < len(blocks); i++ {
			sub := blocks[i]
			//assert sub.floorLeadByte != -1;
			//if (DEBUG) {
			//  System.out.println("    write floorLeadByte=" + Integer.toHexString(sub.floorLeadByte&0xff));
			//}
			if err := scratchBytes.WriteByte(byte(sub.floorLeadByte)); err != nil {
				return err
			}
			//assert sub.fp > fp;
			hasTerm := uint64(0)
			if sub.hasTerms {
				hasTerm = 1
			}
			if err := scratchBytes.WriteUvarint(ctx, uint64(sub.fp-p.fp)<<1|hasTerm); err != nil {
				return err
			}
		}
	}

	estimateSize := len(p.prefix)
	for _, block := range blocks {
		if block.subIndices != nil {
			for _, subIndex := range block.subIndices {
				estimateSize += subIndex.NumBytes()
			}
		}
	}
	estimateBitsRequired, err := packed.BitsRequired(int64(estimateSize))
	if err != nil {
		return err
	}
	pageBits := min(15, max(6, estimateBitsRequired))

	outputs := fst.NewByteSequenceOutputs()
	indexBuilder, err := fst.NewBuilder(fst.BYTE1, outputs,
		fst.WithMinSuffixCount1(0),
		fst.WithMinSuffixCount2(0),
		fst.WithDoShareSuffix(true),
		fst.WithDoShareNonSingletonNodes(false),
		fst.WithShareMaxTailLength(math.MaxInt32),
		fst.WithAllowFixedLengthArcs(true),
		fst.WithBytesPageBits(pageBits),
	)
	if err != nil {
		return err
	}

	buff := make([]byte, scratchBytes.GetFilePointer())
	scratchBytes.WriteTo(buff)
	runes := make([]rune, estimateSize)
	for _, v := range p.prefix {
		runes = append(runes, rune(v))
	}
	if err := indexBuilder.Add(ctx, runes, buff); err != nil {
		return err
	}
	scratchBytes.Reset()

	// Copy over index for all sub-blocks
	for _, block := range blocks {
		if block.subIndices != nil {
			for _, subIndex := range block.subIndices {
				if err := appendIndex(ctx, indexBuilder, subIndex); err != nil {
					return err
				}
			}
			block.subIndices = nil
		}
	}

	newIndex, err := indexBuilder.Finish(ctx)
	if err != nil {
		return err
	}
	p.index = newIndex
	return nil
}

func appendIndex(ctx context.Context, builder *fst.Builder[[]byte], subIndex *fst.FST[[]byte]) error {
	subIndexEnum, err := fst.NewBytesFSTEnum(subIndex)
	if err != nil {
		return err
	}

	for {
		indexEnt, err := subIndexEnum.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		runes := make([]rune, 0, len(indexEnt.GetInput()))
		for _, input := range indexEnt.GetInput() {
			runes = append(runes, rune(input))
		}
		if err := builder.Add(ctx, runes, indexEnt.GetOutput()); err != nil {
			return err
		}
	}
}

func encodeOutput(fp int64, hasTerms, isFloor bool) uint64 {
	res := uint64(0)
	if hasTerms {
		res |= OUTPUT_FLAG_HAS_TERMS
	}
	if isFloor {
		res |= OUTPUT_FLAG_IS_FLOOR
	}
	return res | uint64(fp)<<2
}

type statsWriter struct {
	out            store.DataOutput
	hasFreqs       bool
	singletonCount uint64
}

func newStatsWriter(out store.DataOutput, hasFreqs bool) *statsWriter {
	return &statsWriter{out: out, hasFreqs: hasFreqs}
}

func (s *statsWriter) add(ctx context.Context, df int, ttf int64) error {
	// Singletons (DF==1, TTF==1) are run-length encoded
	if df == 1 && (s.hasFreqs == false || ttf == 1) {
		s.singletonCount++
		return nil
	}

	if err := s.finish(ctx); err != nil {
		return err
	}
	if err := s.out.WriteUvarint(ctx, uint64(df)<<1); err != nil {
		return err
	}
	if s.hasFreqs {
		if err := s.out.WriteUvarint(ctx, uint64(ttf)-uint64(df)); err != nil {
			return err
		}
	}
	return nil
}

func (s *statsWriter) finish(ctx context.Context) error {
	if s.singletonCount > 0 {
		if err := s.out.WriteUvarint(ctx, ((s.singletonCount-1)<<1)|1); err != nil {
			return err
		}
		s.singletonCount = 0
	}
	return nil
}

type termsWriter struct {
	parent *TermsWriter

	fieldInfo        *document.FieldInfo
	numTerms         int
	docsSeen         *bitset.BitSet
	sumTotalTermFreq int
	sumDocFreq       int

	// Records index into pending where the current prefix at that
	// length "started"; for example, if current term starts with 't',
	// startsByPrefix[0] is the index into pending for the first
	// term/sub-block starting with 't'.  We use this to figure out when
	// to write a new block:
	lastTerm     *bytes.Buffer
	prefixStarts []int

	// Pending stack of terms and blocks.  As terms arrive (in sorted order)
	// we append to this stack, and once the top of the stack has enough
	// terms starting with a common prefix, we write a new block with
	// those terms and replace those terms in the stack with a new block:
	pending []pendingEntry

	// Reused in writeBlocks:
	newBlocks []*pendingBlock

	firstPendingTerm *pendingTerm
	lastPendingTerm  *pendingTerm

	statsWriter         *store.RAMOutputStream
	suffixLengthsWriter *store.RAMOutputStream
	metaWriter          *store.RAMOutputStream
	spareWriter         *store.RAMOutputStream
	suffixWriter        *bytes.Buffer

	spareBytes []byte
}

func (w *termsWriter) Write(ctx context.Context, text []byte,
	termsEnum coreIndex.TermsEnum, norms coreIndex.NormsProducer) error {

	state, err := w.parent.postingsWriter.WriteTerm(ctx, text, termsEnum, w.docsSeen, norms)
	if err != nil {
		return err
	}

	if err := w.pushTerm(ctx, text); err != nil {
		return err
	}

	term := newPendingTerm(text, state)
	w.pending = append(w.pending, term)

	w.sumDocFreq += state.GetDocFreq()
	w.sumTotalTermFreq += state.GetTotalTermFreq()
	w.numTerms++
	if w.firstPendingTerm == nil {
		w.firstPendingTerm = term
	}
	w.lastPendingTerm = term
	return nil
}

func (w *termsWriter) pushTerm(ctx context.Context, text []byte) error {
	// Find common prefix between last term and current term:

	prefixLength := array.Mismatch(w.lastTerm.Bytes(), text)
	if prefixLength == -1 { // Only happens for the first term, if it is empty
		prefixLength = 0
	}

	// if (DEBUG) System.out.println("  shared=" + pos + "  lastTerm.length=" + lastTerm.length);

	// Close the "abandoned" suffix now:
	for i := w.lastTerm.Len() - 1; i >= prefixLength; i-- {

		// How many items on top of the stack share the current suffix
		// we are closing:
		prefixTopSize := len(w.pending) - w.prefixStarts[i]
		if prefixTopSize >= w.parent.minItemsInBlock {
			// if (DEBUG) System.out.println("pushTerm i=" + i + " prefixTopSize=" + prefixTopSize + " minItemsInBlock=" + minItemsInBlock);
			if err := w.writeBlocks(ctx, i+1, prefixTopSize); err != nil {
				return err
			}
			w.prefixStarts[i] -= prefixTopSize - 1
		}
	}

	if len(w.prefixStarts) < len(text) {
		w.prefixStarts = array.Grow(w.prefixStarts, len(text))
	}

	// Init new tail:
	for i := prefixLength; i < len(text); i++ {
		w.prefixStarts[i] = len(w.pending)
	}

	w.lastTerm.Reset()
	w.lastTerm.Write(text)
	return nil
}

func (w *termsWriter) writeBlocks(ctx context.Context, prefixLength int, count int) error {

	//assert count > 0;

	//if (DEBUG2) {
	//  BytesRef br = new BytesRef(lastTerm.bytes());
	//  br.length = prefixLength;
	//  System.out.println("writeBlocks: seg=" + segment + " prefix=" + brToString(br) + " count=" + count);
	//}

	// Root block better write all remaining pending entries:
	//assert prefixLength > 0 || count == pending.size();

	lastSuffixLeadLabel := -1

	// True if we saw at least one term in this block (we record if a block
	// only points to sub-blocks in the terms index so we can avoid seeking
	// to it when we are looking for a term):
	hasTerms := false
	hasSubBlocks := false

	start := len(w.pending) - count
	end := len(w.pending)
	nextBlockStart := start
	nextFloorLeadLabel := -1

	for i := start; i < end; i++ {

		ent := w.pending[i]

		var suffixLeadLabel int

		if ent.IsTerm() {
			term := ent.(*pendingTerm)
			if len(term.termBytes) == prefixLength {
				// Suffix is 0, i.e. prefix 'foo' and term is
				// 'foo' so the term has empty string suffix
				// in this block
				//assert lastSuffixLeadLabel == -1: "i=" + i + " lastSuffixLeadLabel=" + lastSuffixLeadLabel;
				suffixLeadLabel = -1
			} else {
				suffixLeadLabel = int(term.termBytes[prefixLength])
			}
		} else {
			block := ent.(*pendingBlock)
			//assert block.prefix.length > prefixLength;
			suffixLeadLabel = int(block.prefix[prefixLength])
		}
		// if (DEBUG) System.out.println("  i=" + i + " ent=" + ent + " suffixLeadLabel=" + suffixLeadLabel);

		if suffixLeadLabel != lastSuffixLeadLabel {
			itemsInBlock := i - nextBlockStart
			if itemsInBlock >= w.parent.minItemsInBlock && end-nextBlockStart > w.parent.maxItemsInBlock {
				// The count is too large for one block, so we must break it into "floor" blocks, where we record
				// the leading label of the suffix of the first term in each floor block, so at search time we can
				// jump to the right floor block.  We just use a naive greedy segmenter here: make a new floor
				// block as soon as we have at least minItemsInBlock.  This is not always best: it often produces
				// a too-small block as the final block:
				isFloor := itemsInBlock < count

				block, err := w.writeBlock(ctx, prefixLength, isFloor, nextFloorLeadLabel, nextBlockStart, i, hasTerms, hasSubBlocks)
				if err != nil {
					return err
				}
				w.newBlocks = append(w.newBlocks, block)

				hasTerms = false
				hasSubBlocks = false
				nextFloorLeadLabel = suffixLeadLabel
				nextBlockStart = i
			}

			lastSuffixLeadLabel = suffixLeadLabel
		}

		if ent.IsTerm() {
			hasTerms = true
		} else {
			hasSubBlocks = true
		}
	}

	// Write last block, if any:
	if nextBlockStart < end {
		itemsInBlock := end - nextBlockStart
		isFloor := itemsInBlock < count

		block, err := w.writeBlock(ctx, prefixLength, isFloor, nextFloorLeadLabel, nextBlockStart, end, hasTerms, hasSubBlocks)
		if err != nil {
			return err
		}

		w.newBlocks = append(w.newBlocks, block)
	}

	//assert newBlocks.isEmpty() == false;

	firstBlock := w.newBlocks[0]

	//assert firstBlock.isFloor || newBlocks.size() == 1;

	if err := firstBlock.compileIndex(ctx, w.newBlocks, w.parent.scratchBytes); err != nil {
		return err
	}

	// Remove slice from the top of the pending stack, that we just wrote:
	w.pending = w.pending[:count]

	// Append new block
	w.pending = append(w.pending, firstBlock)

	w.newBlocks = w.newBlocks[0:0]

	return nil
}

func (w *termsWriter) writeBlock(ctx context.Context, prefixLength int, isFloor bool, floorLeadLabel, start, end int,
	hasTerms, hasSubBlocks bool) (*pendingBlock, error) {

	startFP := w.parent.termsOut.GetFilePointer()

	hasFloorLeadLabel := isFloor && floorLeadLabel != -1

	var prefix []byte
	if hasFloorLeadLabel {
		prefix = make([]byte, prefixLength+1)
	} else {
		prefix = make([]byte, prefixLength)
	}
	copy(prefix, w.lastTerm.Bytes()[:prefixLength])
	prefix = prefix[:prefixLength]

	// Write block header:
	numEntries := end - start
	code := numEntries << 1
	if end == len(w.pending) {
		// Last block:
		code |= 1
	}
	if err := w.parent.termsOut.WriteUvarint(ctx, uint64(code)); err != nil {
		return nil, err
	}

	// We optimize the leaf block case (block has only terms), writing a more
	// compact format in this case:
	isLeafBlock := hasSubBlocks == false

	subIndices := make([]*fst.FST[[]byte], 0)
	absolute := true

	if isLeafBlock {
		// Block contains only ordinary terms:
		subIndices = nil
		hasFreq := w.fieldInfo.GetIndexOptions() != document.INDEX_OPTIONS_DOCS
		sw := newStatsWriter(w.statsWriter, hasFreq)
		for i := start; i < end; i++ {
			ent := w.pending[i]

			term := ent.(*pendingTerm)

			//assert StringHelper.startsWith(term.termBytes, prefix): "term.term=" + term.termBytes + " prefix=" + prefix;
			state := term.state
			suffix := len(term.termBytes) - prefixLength
			//if (DEBUG2) {
			//  BytesRef suffixBytes = new BytesRef(suffix);
			//  System.arraycopy(term.termBytes, prefixLength, suffixBytes.bytes, 0, suffix);
			//  suffixBytes.length = suffix;
			//  System.out.println("    write term suffix=" + brToString(suffixBytes));
			//}

			// For leaf block we write suffix straight
			if err := w.suffixLengthsWriter.WriteUvarint(ctx, uint64(suffix)); err != nil {
				return nil, err
			}
			if _, err := w.suffixWriter.Write(term.termBytes[prefixLength : prefixLength+suffix]); err != nil {
				return nil, err
			}

			// Write term stats, to separate byte[] blob:
			if err := sw.add(ctx, state.GetDocFreq(), int64(state.GetTotalTermFreq())); err != nil {
				return nil, err
			}

			// Write term meta data
			if err := w.parent.postingsWriter.EncodeTerm(ctx, w.metaWriter, w.fieldInfo, state, absolute); err != nil {
				return nil, err
			}
			absolute = false
		}
		if err := sw.finish(ctx); err != nil {
			return nil, err
		}
	} else {
		// Block has at least one prefix term or a sub block:
		subIndices = make([]*fst.FST[[]byte], 0)
		writer := newStatsWriter(w.statsWriter, w.fieldInfo.GetIndexOptions() != document.INDEX_OPTIONS_DOCS)
		for i := start; i < end; i++ {
			ent := w.pending[i]
			if ent.IsTerm() {
				term := ent.(*pendingTerm)

				//assert StringHelper.startsWith(term.termBytes, prefix): "term.term=" + term.termBytes + " prefix=" + prefix;
				state := term.state
				suffix := len(term.termBytes) - prefixLength
				//if (DEBUG2) {
				//  BytesRef suffixBytes = new BytesRef(suffix);
				//  System.arraycopy(term.termBytes, prefixLength, suffixBytes.bytes, 0, suffix);
				//  suffixBytes.length = suffix;
				//  System.out.println("      write term suffix=" + brToString(suffixBytes));
				//}

				// For non-leaf block we borrow 1 bit to record
				// if entry is term or sub-block, and 1 bit to record if
				// it's a prefix term.  Terms cannot be larger than ~32 KB
				// so we won't run out of bits:

				if err := w.suffixLengthsWriter.WriteUvarint(ctx, uint64(suffix<<1)); err != nil {
					return nil, err
				}
				w.suffixWriter.Write(term.termBytes[prefixLength : prefixLength+suffix])

				// Write term stats, to separate byte[] blob:
				if err := writer.add(ctx, state.GetDocFreq(), int64(state.GetTotalTermFreq())); err != nil {
					return nil, err
				}

				// TODO: now that terms dict "sees" these longs,
				// we can explore better column-stride encodings
				// to encode all long[0]s for this block at
				// once, all long[1]s, etc., e.g. using
				// Simple64.  Alternatively, we could interleave
				// stats + meta ... no reason to have them
				// separate anymore:

				// Write term meta data
				if err := w.parent.postingsWriter.EncodeTerm(ctx,
					w.metaWriter, w.fieldInfo, state, absolute); err != nil {
					return nil, err
				}
				absolute = false
			} else {
				block := ent.(*pendingBlock)
				suffix := len(block.prefix) - prefixLength

				// For non-leaf block we borrow 1 bit to record
				// if entry is term or sub-block:f
				if err := w.suffixLengthsWriter.WriteUvarint(ctx, uint64(suffix<<1)|1); err != nil {
					return nil, err
				}
				w.suffixWriter.Write(block.prefix[prefixLength : prefixLength+suffix])

				//if (DEBUG2) {
				//  BytesRef suffixBytes = new BytesRef(suffix);
				//  System.arraycopy(block.prefix.bytes, prefixLength, suffixBytes.bytes, 0, suffix);
				//  suffixBytes.length = suffix;
				//  System.out.println("      write sub-block suffix=" + brToString(suffixBytes) + " subFP=" + block.fp + " subCode=" + (startFP-block.fp) + " floor=" + block.isFloor);
				//}

				if err := w.suffixLengthsWriter.WriteUvarint(ctx, uint64(startFP-block.fp)); err != nil {
					return nil, err
				}
				subIndices = append(subIndices, block.index)
			}
		}
		if err := writer.finish(ctx); err != nil {
			return nil, err
		}

	}

	// Write suffixes byte[] blob to terms dict output, either uncompressed,
	// compressed with LZ4 or with LowercaseAsciiCompression.
	var compressionAlg CompressionAlgorithm

	// If there are 2 suffix bytes or less per term, then we don't bother compressing as suffix are unlikely what
	// makes the terms dictionary large, and it also tends to be frequently the case for dense IDs like
	// auto-increment IDs, so not compressing in that case helps not hurt ID lookups by too much.
	// We also only start compressing when the prefix length is greater than 2 since blocks whose prefix length is
	// 1 or 2 always all get visited when running a fuzzy query whose max number of edits is 2.
	if w.suffixWriter.Len() > 2*numEntries && prefixLength > 2 {
		// LZ4 inserts references whenever it sees duplicate strings of 4 chars or more, so only try it out if the
		// average suffix length is greater than 6.
		if w.suffixWriter.Len() > 6*numEntries {
			if err := compress.LZ4Compression.Compress(w.suffixWriter.Bytes(), w.spareWriter); err != nil {
				return nil, err
			}
			if int(w.spareWriter.GetFilePointer()) < w.suffixWriter.Len()-(w.suffixWriter.Len()>>2) {
				// LZ4 saved more than 25%, go for it
				compressionAlg = LZ4
			}
		}
		if _, ok := compressionAlg.(*noCompression); ok {
			w.spareWriter.Reset()
			if len(w.spareBytes) < w.suffixWriter.Len() {
				w.spareBytes = make([]byte, array.Oversize(w.suffixWriter.Len(), 1))
			}

			ok, err := compress.LowercaseAsciiCompression.Compress(ctx,
				w.suffixWriter.Bytes(), w.spareBytes, w.spareWriter)
			if err != nil {
				return nil, err
			}
			if ok {
				compressionAlg = LOWERCASE_ASCII
			}
		}
	}
	token := (w.suffixWriter.Len()) << 3
	if isLeafBlock {
		token |= 0x04
	}
	token |= compressionAlg.Code()
	if err := w.parent.termsOut.WriteUvarint(ctx, uint64(token)); err != nil {
		return nil, err
	}
	if _, ok := compressionAlg.(*noCompression); ok {
		if _, err := w.parent.termsOut.Write(w.suffixWriter.Bytes()); err != nil {
			return nil, err
		}
	} else {
		if err := w.spareWriter.WriteToDataOutput(w.parent.termsOut); err != nil {
			return nil, err
		}
	}
	w.suffixWriter.Reset()
	w.spareWriter.Reset()

	// Write suffix lengths
	numSuffixBytes := int(w.suffixLengthsWriter.GetFilePointer())
	w.spareBytes = array.Grow(w.spareBytes, numSuffixBytes)
	w.suffixLengthsWriter.WriteTo(w.spareBytes)
	w.suffixLengthsWriter.Reset()
	if allEqual(w.spareBytes[1:1+numSuffixBytes], w.spareBytes[0]) {
		// Structured fields like IDs often have most values of the same length
		if err := w.parent.termsOut.WriteUvarint(ctx, uint64(numSuffixBytes<<1)|1); err != nil {
			return nil, err
		}
		if err := w.parent.termsOut.WriteByte(w.spareBytes[0]); err != nil {
			return nil, err
		}
	} else {
		if err := w.parent.termsOut.WriteUvarint(ctx, uint64(numSuffixBytes<<1)); err != nil {
			return nil, err
		}
		if _, err := w.parent.termsOut.Write(w.spareBytes[:numSuffixBytes]); err != nil {
			return nil, err
		}
	}

	// Stats
	numStatsBytes := int(w.statsWriter.GetFilePointer())
	if err := w.parent.termsOut.WriteUvarint(ctx, uint64(numStatsBytes)); err != nil {
		return nil, err
	}
	if err := w.statsWriter.WriteToDataOutput(w.parent.termsOut); err != nil {
		return nil, err
	}
	w.statsWriter.Reset()

	// Write term meta data byte[] blob
	if err := w.parent.termsOut.WriteUvarint(ctx, uint64(w.metaWriter.GetFilePointer())); err != nil {
		return nil, err
	}
	if err := w.metaWriter.WriteToDataOutput(w.parent.termsOut); err != nil {
		return nil, err
	}
	w.metaWriter.Reset()

	if hasFloorLeadLabel {
		// We already allocated to length+1 above:
		prefix = append(prefix, byte(floorLeadLabel))
	}
	return newPendingBlock(prefix, startFP, hasTerms, isFloor, floorLeadLabel, subIndices), nil
}

func allEqual(bs []byte, b byte) bool {
	for _, v := range bs {
		if v != b {
			return false
		}
	}
	return true
}

func (w *termsWriter) Finish(ctx context.Context) error {
	if w.numTerms > 0 {
		// if (DEBUG) System.out.println("BTTW: finish prefixStarts=" + Arrays.toString(prefixStarts));

		// Add empty term to force closing of all final blocks:
		if err := w.pushTerm(ctx, []byte{}); err != nil {
			return err
		}

		// TODO: if pending.size() is already 1 with a non-zero prefix length
		// we can save writing a "degenerate" root block, but we have to
		// fix all the places that assume the root block's prefix is the empty string:
		if err := w.pushTerm(ctx, []byte{}); err != nil {
			return err
		}
		if err := w.writeBlocks(ctx, 0, len(w.pending)); err != nil {
			return err
		}

		// We better have one final "root" block:
		root := w.pending[0].(*pendingBlock)
		rootCode := root.index.GetEmptyOutput()

		metaOut := store.NewBufferDataOutput()
		w.parent.fields = append(w.parent.fields, metaOut)

		if err := metaOut.WriteUvarint(ctx, uint64(w.fieldInfo.Number())); err != nil {
			return err
		}
		if err := metaOut.WriteUvarint(ctx, uint64(w.numTerms)); err != nil {
			return err
		}
		if err := metaOut.WriteUvarint(ctx, uint64(len(rootCode))); err != nil {
			return err
		}
		if _, err := metaOut.Write(rootCode); err != nil {
			return err
		}
		if w.fieldInfo.GetIndexOptions() != document.INDEX_OPTIONS_DOCS {
			if err := metaOut.WriteUvarint(ctx, uint64(w.sumTotalTermFreq)); err != nil {
				return err
			}
		}
		if err := metaOut.WriteUvarint(ctx, uint64(w.sumDocFreq)); err != nil {
			return err
		}
		if err := metaOut.WriteUvarint(ctx, uint64(w.docsSeen.Count())); err != nil {
			return err
		}
		if err := writeBytes(ctx, metaOut, w.firstPendingTerm.termBytes); err != nil {
			return err
		}
		if err := writeBytes(ctx, metaOut, w.lastPendingTerm.termBytes); err != nil {
			return err
		}
		if err := metaOut.WriteUvarint(ctx, uint64(w.parent.indexOut.GetFilePointer())); err != nil {
			return err
		}
		// Write FST to index
		if err := root.index.Save(ctx, metaOut, w.parent.indexOut); err != nil {
			return err
		}

	}
	return nil
}

func writeBytes(ctx context.Context, out store.DataOutput, termBytes []byte) error {
	if err := out.WriteUvarint(ctx, uint64(len(termBytes))); err != nil {
		return err
	}
	if _, err := out.Write(termBytes); err != nil {
		return err
	}
	return nil
}

func (t *TermsWriter) newTermsWriter(fieldInfo *document.FieldInfo) *termsWriter {
	t.postingsWriter.SetField(fieldInfo)
	tw := &termsWriter{
		parent: t,

		fieldInfo:           fieldInfo,
		docsSeen:            bitset.New(uint(t.maxDoc)),
		statsWriter:         store.NewRAMOutputStream("noname", store.NewRAMFile(nil), false),
		suffixLengthsWriter: store.NewRAMOutputStream("noname", store.NewRAMFile(nil), false),
		metaWriter:          store.NewRAMOutputStream("noname", store.NewRAMFile(nil), false),
		spareWriter:         store.NewRAMOutputStream("noname", store.NewRAMFile(nil), false),
		suffixWriter:        new(bytes.Buffer),
	}
	return tw
}
