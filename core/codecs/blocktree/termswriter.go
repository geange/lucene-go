package blocktree

import (
	"bytes"
	"context"

	"github.com/bits-and-blooms/bitset"

	"github.com/geange/lucene-go/core/codecs/types"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	coreIndex "github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/fst"
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

	panic("")
}

func (t *TermsWriter) Close() error {
	//TODO implement me
	panic("implement me")
}

func (t *TermsWriter) Write(ctx context.Context, fields coreIndex.Fields, norms coreIndex.NormsProducer) error {
	//TODO implement me
	panic("implement me")
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

func (p *pendingTerm) IsTerm() bool {
	return p.isTerm
}

var _ pendingEntry = &pendingBlock{}

type pendingBlock struct {
	prefix        []byte
	fp            int64
	index         *fst.FST
	subIndices    []*fst.FST
	hasTerms      bool
	isFloor       bool
	floorLeadByte int
	isTerm        bool
}

func (p *pendingBlock) IsTerm() bool {
	return p.isTerm
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
	newBlocks []pendingBlock

	firstPendingTerm pendingTerm
	lastPendingTerm  pendingTerm
}
