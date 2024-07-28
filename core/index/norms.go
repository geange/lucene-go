package index

import (
	"context"
	"errors"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/util/packed"
	"io"

	"github.com/geange/lucene-go/core/document"
)

// NormsConsumer
// Abstract API that consumes normalization values. Concrete implementations of this actually do "something"
// with the norms (write it into the index in a specific format).
// The lifecycle is:
// NormsConsumer is created by NormsFormat.normsConsumer(SegmentWriteState).
// FnAddNormsField is called for each field with normalization values. The API is a "pull" rather than "push",
// and the implementation is free to iterate over the values multiple times (Iterable.iterator()).
// After all fields are added, the consumer is closed.
type NormsConsumer interface {
	io.Closer

	// AddNormsField
	// Writes normalization values for a field.
	// field: field information
	// normsProducer: NormsProducer of the numeric norm values
	// Throws: IOException – if an I/O error occurred.
	AddNormsField(ctx context.Context, field *document.FieldInfo, normsProducer index.NormsProducer) error

	// Merge
	// Merges in the fields from the readers in mergeState.
	// The default implementation calls mergeNormsField for each field,
	// filling segments with missing norms for the field with zeros.
	// Implementations can override this method for more sophisticated merging
	// (bulk-byte copying, etc).
	Merge(ctx context.Context, mergeState *index.MergeState) error

	// MergeNormsField
	// Merges the norms from toMerge.
	// The default implementation calls FnAddNormsField, passing an Iterable
	// that merges and filters deleted documents on the fly.
	MergeNormsField(ctx context.Context, mergeFieldInfo *document.FieldInfo, mergeState *index.MergeState) error
}

type NormsConsumerDefault struct {
	FnAddNormsField func(ctx context.Context, field *document.FieldInfo, normsProducer index.NormsProducer) error
}

func (n *NormsConsumerDefault) Merge(ctx context.Context, mergeState *index.MergeState) error {
	for _, normsProducer := range mergeState.NormsProducers {
		if normsProducer != nil {
			if err := normsProducer.CheckIntegrity(); err != nil {
				return err
			}
		}
	}
	for _, mergeFieldInfo := range mergeState.MergeFieldInfos.List() {
		if mergeFieldInfo.HasNorms() {
			if err := n.MergeNormsField(ctx, mergeFieldInfo, mergeState); err != nil {
				return err
			}
		}
	}
	return nil
}

func (n *NormsConsumerDefault) MergeNormsField(ctx context.Context, mergeFieldInfo *document.FieldInfo, mergeState *index.MergeState) error {
	// TODO: try to share code with default merge of DVConsumer by passing MatchAllBits ?
	return n.FnAddNormsField(ctx, mergeFieldInfo, &innerNormsProducer{
		mergeFieldInfo: mergeFieldInfo,
		mergeState:     mergeState,
	})
}

var _ index.NormsProducer = &innerNormsProducer{}

type innerNormsProducer struct {
	mergeFieldInfo *document.FieldInfo
	mergeState     *index.MergeState
}

func (i *innerNormsProducer) Close() error {
	return nil
}

func (i *innerNormsProducer) GetNorms(fieldInfo *document.FieldInfo) (index.NumericDocValues, error) {
	if fieldInfo != i.mergeFieldInfo {
		return nil, errors.New("wrong fieldInfo")
	}

	// TODO: impl it

	//subs :=
	//List<NumericDocValuesSub> subs = new ArrayList<>();
	//assert mergeState.docMaps.length == mergeState.docValuesProducers.length;
	//for (int i=0;i<mergeState.docValuesProducers.length;i++) {
	//	NumericDocValues norms = null;
	//	NormsProducer normsProducer = mergeState.normsProducers[i];
	//	if (normsProducer != null) {
	//		FieldInfo readerFieldInfo = mergeState.fieldInfos[i].fieldInfo(mergeFieldInfo.name);
	//		if (readerFieldInfo != null && readerFieldInfo.hasNorms()) {
	//			norms = normsProducer.getNorms(readerFieldInfo);
	//		}
	//	}
	//
	//	if (norms != null) {
	//		subs.add(new NumericDocValuesSub(mergeState.docMaps[i], norms));
	//	}
	//}
	//
	//final DocIDMerger<NumericDocValuesSub> docIDMerger = DocIDMerger.of(subs, mergeState.needsIndexSort);

	panic("")
}

func (i *innerNormsProducer) CheckIntegrity() error {
	return nil
}

func (i *innerNormsProducer) GetMergeInstance() index.NormsProducer {
	return i
}

// NumericDocValuesSub Tracks state of one numeric sub-reader that we are merging
type NumericDocValuesSub struct {
	values index.NumericDocValues
}

// NormValuesWriter
// Buffers up pending long per doc, then flushes when segment flushes.
type NormValuesWriter struct {
	docsWithField *DocsWithFieldSet
	pending       *packed.PackedLongValuesBuilder
	fieldInfo     *document.FieldInfo
	lastDocID     int
}

func (n *NormValuesWriter) AddValue(docID int, value int64) error {
	if n.lastDocID >= docID {
		return errors.New("docID too small")
	}

	if err := n.pending.Add(value); err != nil {
		return err
	}

	n.lastDocID = docID
	return n.docsWithField.Add(docID)
}

func (n *NormValuesWriter) Finish(maxDoc int) {

}

func (n *NormValuesWriter) Flush(state *index.SegmentWriteState, sortMap *DocMap, normsConsumer index.NormsConsumer) error {
	//values := n.pending.Build()
	// TODO: impl it
	panic("")
}

func NewNormValuesWriter(fieldInfo *document.FieldInfo) *NormValuesWriter {
	//return &NormValuesWriter{
	//	docsWithField: NewDocsWithFieldSet(),
	//	pending:       packed.NewLongValuesBuilder(make([]uint64, 0)...),
	//	fieldInfo:     fieldInfo,
	//	lastDocID:     -1,
	//}
	// TODO: impl it
	panic("")
}
