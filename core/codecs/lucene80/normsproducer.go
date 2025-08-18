package lucene80

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/geange/lucene-go/core/codecs"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.NormsProducer = &NormsProducer{}

type NormsProducer struct {
	// metadata maps (just file pointers and minimal stuff)
	//private final Map<Integer,NormsEntry> norms = new HashMap<>();
	norms          map[int]*NormsEntry
	maxDoc         int
	data           store.IndexInput
	merging        bool
	disiInputs     map[int]store.IndexInput
	disiJumpTables map[int]*store.RandomAccessInput
	dataInputs     map[int]*store.RandomAccessInput
}

func NewNormsProducer(ctx context.Context, state *index.SegmentReadState,
	dataCodec, dataExtension, metaCodec, metaExtension string) (*NormsProducer, error) {

	producer := &NormsProducer{
		norms: make(map[int]*NormsEntry),
	}

	maxDoc, err := state.SegmentInfo.MaxDoc()
	if err != nil {
		return nil, err
	}
	producer.maxDoc = maxDoc

	metaName := store.SegmentFileName(state.SegmentInfo.Name(), state.SegmentSuffix, metaExtension)
	version := -1

	in, err := store.OpenChecksumInput(ctx, state.Directory, metaName)
	if err != nil {
		return nil, err
	}
	headerVersion, err := codecs.CheckIndexHeader(ctx, in, metaCodec,
		NORMS_VERSION_START, NORMS_VERSION_CURRENT, state.SegmentInfo.GetId(), state.SegmentSuffix)
	if err != nil {
		return nil, err
	}
	version = headerVersion

	if err := producer.readFields(ctx, in, state.FieldInfos); err != nil {
		return nil, err
	}
	if _, err := codecs.CheckFooter(ctx, in); err != nil {
		return nil, err
	}

	dataName := store.SegmentFileName(state.SegmentInfo.Name(), state.SegmentSuffix, dataExtension)
	data, err := state.Directory.OpenInput(ctx, dataName)
	if err != nil {
		return nil, err
	}
	producer.data = data

	version2, err := codecs.CheckIndexHeader(ctx, data, dataCodec,
		NORMS_VERSION_START, NORMS_VERSION_CURRENT, state.SegmentInfo.GetId(), state.SegmentSuffix)
	if err != nil {
		return nil, err
	}

	if version != version2 {
		return nil, fmt.Errorf("format versions mismatch: meta=%d, data=%d", version, version2)
	}

	// NOTE: data file is too costly to verify checksum against all the bytes on open,
	// but for now we at least verify proper structure of the checksum footer: which looks
	// for FOOTER_MAGIC + algorithmID. This is cheap and can detect some forms of corruption
	// such as file truncation.
	if _, err := codecs.RetrieveChecksum(ctx, data); err != nil {
		return nil, err
	}

	return producer, nil
}

func (n *NormsProducer) readFields(ctx context.Context, meta store.ChecksumIndexInput, infos index.FieldInfos) error {
	for {
		fieldNumber, err := meta.ReadUint32(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		if int32(fieldNumber) == -1 {
			break
		}

		info := infos.FieldInfoByNumber(int(fieldNumber))
		if info == nil {
			return fmt.Errorf("invalid field number: %d", fieldNumber)
		}
		if !info.HasNorms() {
			return fmt.Errorf("invalid field: %s", info.Name())
		}

		entry := &NormsEntry{}
		docsWithFieldOffset, err := meta.ReadUint64(ctx)
		if err != nil {
			return err
		}
		entry.docsWithFieldOffset = int64(docsWithFieldOffset)

		docsWithFieldLength, err := meta.ReadUint64(ctx)
		if err != nil {
			return err
		}
		entry.docsWithFieldLength = int(docsWithFieldLength)

		jumpTableEntryCount, err := meta.ReadByte()
		if err != nil {
			return err
		}
		entry.jumpTableEntryCount = int(jumpTableEntryCount)

		denseRankPower, err := meta.ReadByte()
		if err != nil {
			return err
		}
		entry.denseRankPower = denseRankPower

		numDocsWithField, err := meta.ReadUint32(ctx)
		if err != nil {
			return err
		}
		entry.numDocsWithField = int(int32(numDocsWithField))

		bytesPerNorm, err := meta.ReadByte()
		if err != nil {
			return err
		}
		entry.bytesPerNorm = bytesPerNorm

		switch entry.bytesPerNorm {
		case 0, 1, 2, 4, 8:
		default:
			return fmt.Errorf("invalid bytesPerValue: %d, field: %s", entry.bytesPerNorm, info.Name())
		}

		normsOffset, err := meta.ReadUint64(ctx)
		if err != nil {
			return err
		}
		entry.normsOffset = int64(normsOffset)
		n.norms[info.Number()] = entry
	}
	return nil
}

func (n *NormsProducer) Close() error {
	//TODO implement me
	panic("implement me")
}

func (n *NormsProducer) GetNorms(field *document.FieldInfo) (index.NumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (n *NormsProducer) CheckIntegrity() error {
	_, err := codecs.ChecksumEntireFile(context.Background(), n.data)
	return err
}

func (n *NormsProducer) GetMergeInstance() index.NormsProducer {
	//TODO implement me
	panic("implement me")
}

type NormsEntry struct {
	denseRankPower      byte
	bytesPerNorm        byte
	docsWithFieldOffset int64
	docsWithFieldLength int
	jumpTableEntryCount int
	numDocsWithField    int
	normsOffset         int64
}

type DenseNormsIterator struct {
	maxDoc int
	doc    int
}

func NewDenseNormsIterator(maxDoc int) *DenseNormsIterator {
	return &DenseNormsIterator{maxDoc: maxDoc}
}

func (d *DenseNormsIterator) DocID() int {
	return d.doc
}

func (d *DenseNormsIterator) NextDoc(ctx context.Context) (int, error) {
	return d.Advance(ctx, d.doc+1)
}

func (d *DenseNormsIterator) Advance(ctx context.Context, target int) (int, error) {
	if target >= d.maxDoc {
		return -1, io.EOF
	}
	d.doc = target
	return target, nil
}

func (d *DenseNormsIterator) Cost() int64 {
	return int64(d.maxDoc)
}

func (d *DenseNormsIterator) AdvanceExact(target int) (bool, error) {
	d.doc = target
	return true, nil
}

type SparseNormsIterator struct {
	disi *IndexedDISI
}

func NewSparseNormsIterator(disi *IndexedDISI) *SparseNormsIterator {
	return &SparseNormsIterator{disi: disi}
}

func (s *SparseNormsIterator) DocID() int {
	return s.disi.DocID()
}

func (s *SparseNormsIterator) NextDoc(ctx context.Context) (int, error) {
	return s.disi.NextDoc(ctx)
}

func (s *SparseNormsIterator) Advance(ctx context.Context, target int) (int, error) {
	return s.disi.Advance(ctx, target)
}

func (s *SparseNormsIterator) Cost() int64 {
	return s.disi.Cost()
}

func (s *SparseNormsIterator) AdvanceExact(target int) (bool, error) {
	return s.disi.AdvanceExact(context.Background(), target)
}
