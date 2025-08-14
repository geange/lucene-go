package lucene80

import (
	"context"

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
	panic("implement me")
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
	//TODO implement me
	panic("implement me")
}

func (n *NormsProducer) GetMergeInstance() index.NormsProducer {
	//TODO implement me
	panic("implement me")
}

type NormsEntry struct {
	denseRankPower      byte
	bytesPerNorm        byte
	docsWithFieldOffset int64
	docsWithFieldLength int64
	jumpTableEntryCount int
	numDocsWithField    int
	normsOffset         int64
}
