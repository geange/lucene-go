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
	"github.com/geange/lucene-go/core/types"
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
	disiJumpTables map[int]store.RandomAccessInput
	dataInputs     map[int]store.RandomAccessInput
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
	entry, ok := n.norms[field.Number()]
	if !ok {
		return nil, fmt.Errorf("invalid field:%s number: %d", field.Name(), field.Number())
	}
	if entry.docsWithFieldOffset == -2 {
		// empty
		return NewEmptyNumeric(), nil
	}

	if entry.docsWithFieldOffset == -1 {
		// dense
		if entry.bytesPerNorm == 0 {
			fn := func(d *DenseNormsIterator) (int64, error) {
				return entry.normsOffset, nil
			}

			return NewDenseNormsIterator(n.maxDoc, fn), nil
		}
		slice, err := n.getDataInput(field, entry)
		if err != nil {
			return nil, err
		}
		switch entry.bytesPerNorm {
		case 1:
			fn := func(d *DenseNormsIterator) (int64, error) {
				v, err := slice.ReadU8(int64(d.doc))
				if err != nil {
					return 0, err
				}
				return int64(v), nil
			}

			return NewDenseNormsIterator(n.maxDoc, fn), nil

		case 2:
			fn := func(d *DenseNormsIterator) (int64, error) {
				v, err := slice.ReadU16(int64(d.doc) << 1)
				if err != nil {
					return 0, err
				}
				return int64(v), nil
			}
			return NewDenseNormsIterator(n.maxDoc, fn), nil
		case 4:
			fn := func(d *DenseNormsIterator) (int64, error) {
				v, err := slice.ReadU32(int64(d.doc) << 2)
				if err != nil {
					return 0, err
				}
				return int64(v), nil
			}
			return NewDenseNormsIterator(n.maxDoc, fn), nil
		case 8:
			fn := func(d *DenseNormsIterator) (int64, error) {
				v, err := slice.ReadU64(int64(d.doc) << 3)
				if err != nil {
					return 0, err
				}
				return int64(v), nil
			}
			return NewDenseNormsIterator(n.maxDoc, fn), nil
		default:
			// should not happen, we already validate bytesPerNorm in readFields
			return nil, fmt.Errorf("invalid bytesPerValue: %d, field: %s", entry.bytesPerNorm, field.Name())
		}
	}

	// sparse
	disiInput, err := n.getDisiInput(field, entry)
	if err != nil {
		return nil, err
	}
	disiJumpTable, err := n.getDisiJumpTable(field, entry)
	if err != nil {
		return nil, err
	}
	disi, err := newIndexedDISI(disiInput, disiJumpTable, entry.jumpTableEntryCount, int8(entry.denseRankPower))
	if err != nil {
		return nil, err
	}

	if entry.bytesPerNorm == 0 {
		// TODO:
		//return new SparseNormsIterator(disi) {
		//	@Override
		//	public long longValue() throws IOException {
		//		return entry.normsOffset;
		//	}
		//};
	}
	slice, err := n.getDataInput(field, entry)
	if err != nil {
		return nil, err
	}
	switch entry.bytesPerNorm {
	case 1:
		fn := func(d *SparseNormsIterator) (int64, error) {
			v, err := slice.ReadU8(int64(d.disi.Index()))
			if err != nil {
				return 0, err
			}
			return int64(v), nil
		}

		return NewSparseNormsIterator(disi, fn), nil

	case 2:
		fn := func(d *SparseNormsIterator) (int64, error) {
			v, err := slice.ReadU16(int64(d.disi.Index()) << 1)
			if err != nil {
				return 0, err
			}
			return int64(v), nil
		}

		return NewSparseNormsIterator(disi, fn), nil
	case 4:
		fn := func(d *SparseNormsIterator) (int64, error) {
			v, err := slice.ReadU32(int64(d.disi.Index()) << 2)
			if err != nil {
				return 0, err
			}
			return int64(v), nil
		}

		return NewSparseNormsIterator(disi, fn), nil
	case 8:
		fn := func(d *SparseNormsIterator) (int64, error) {
			v, err := slice.ReadU64(int64(d.disi.Index()) << 3)
			if err != nil {
				return 0, err
			}
			return int64(v), nil
		}

		return NewSparseNormsIterator(disi, fn), nil
	default:
		// should not happen, we already validate bytesPerNorm in readFields
		return nil, fmt.Errorf("invalid bytesPerValue: %d, field: %s", entry.bytesPerNorm, field.Name())
	}
}

func (n *NormsProducer) getDataInput(field *document.FieldInfo, entry *NormsEntry) (store.RandomAccessInput, error) {
	var slice store.RandomAccessInput
	if n.merging {
		slice = n.dataInputs[field.Number()]
	}
	if slice == nil {
		var err error
		slice, err = n.data.RandomAccessSlice(entry.normsOffset, int64(entry.numDocsWithField)*int64(entry.bytesPerNorm))
		if err != nil {
			return nil, err
		}
		if n.merging {
			n.dataInputs[field.Number()] = slice
		}
	}
	return slice, nil
}

func (n *NormsProducer) CheckIntegrity() error {
	_, err := codecs.ChecksumEntireFile(context.Background(), n.data)
	return err
}

func (n *NormsProducer) GetMergeInstance() index.NormsProducer {
	//TODO implement me
	panic("implement me")
}

func (n *NormsProducer) getDisiInput(field *document.FieldInfo, entry *NormsEntry) (store.IndexInput, error) {
	panic("implement me")
}

var _ store.IndexInput = &disiInput{}

type disiInput struct {
	*store.BaseIndexInput

	offset int64
}

func (d *disiInput) Read(p []byte) (n int, err error) {
	//TODO implement me
	panic("implement me")
}

func (d *disiInput) Clone() store.CloneReader {
	//TODO implement me
	panic("implement me")
}

func (d *disiInput) Seek(offset int64, whence int) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (d *disiInput) GetFilePointer() int64 {
	//TODO implement me
	panic("implement me")
}

func (d *disiInput) Slice(sliceDescription string, offset, length int64) (store.IndexInput, error) {
	//TODO implement me
	panic("implement me")
}

func (d *disiInput) Length() int64 {
	//TODO implement me
	panic("implement me")
}

func (n *NormsProducer) getDisiJumpTable(field *document.FieldInfo, entry *NormsEntry) (store.RandomAccessInput, error) {
	var jumpTable store.RandomAccessInput
	if n.merging {
		jumpTable = n.disiJumpTables[field.Number()]
	}
	if jumpTable == nil {
		var err error
		jumpTable, err = createJumpTable(n.data, entry.docsWithFieldOffset, entry.docsWithFieldLength, entry.jumpTableEntryCount)
		if err != nil {
			return nil, err
		}
		if n.merging {
			n.disiJumpTables[field.Number()] = jumpTable
		}
	}
	return jumpTable, nil
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

var _ index.NumericDocValues = &DenseNormsIterator{}

type DenseNormsIterator struct {
	maxDoc    int
	doc       int
	longValue longValueFunc
}

type longValueFunc func(d *DenseNormsIterator) (int64, error)

func NewDenseNormsIterator(maxDoc int, fn longValueFunc) *DenseNormsIterator {
	return &DenseNormsIterator{
		maxDoc:    maxDoc,
		doc:       -1,
		longValue: fn,
	}
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

func (d *DenseNormsIterator) SlowAdvance(ctx context.Context, target int) (int, error) {
	return types.SlowAdvanceWithContext(ctx, d, target)
}

func (d *DenseNormsIterator) LongValue() (int64, error) {
	return d.longValue(d)
}

func (d *DenseNormsIterator) AdvanceExact(target int) (bool, error) {
	d.doc = target
	return true, nil
}

var _ index.NumericDocValues = &SparseNormsIterator{}

type SparseNormsIterator struct {
	disi      *IndexedDISI
	longValue func(*SparseNormsIterator) (int64, error)
}

func (s *SparseNormsIterator) SlowAdvance(ctx context.Context, target int) (int, error) {
	return types.SlowAdvanceWithContext(ctx, s, target)
}

func (s *SparseNormsIterator) LongValue() (int64, error) {
	return s.longValue(s)
}

func NewSparseNormsIterator(disi *IndexedDISI, fn func(*SparseNormsIterator) (int64, error)) *SparseNormsIterator {
	return &SparseNormsIterator{
		disi:      disi,
		longValue: fn,
	}
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
