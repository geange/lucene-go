package index

import (
	"context"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/util/bytesref"
	"github.com/geange/lucene-go/core/util/ints"
)

var _ TermsHash = &FreqProxTermsWriter{}

type FreqProxTermsWriter struct {
	*BaseTermsHash
}

func NewFreqProxTermsWriter(intBlockAllocator ints.IntsAllocator,
	byteBlockAllocator bytesref.Allocator, nextTermsHash TermsHash) *FreqProxTermsWriter {

	return &FreqProxTermsWriter{
		NewTermsHashDefault(intBlockAllocator, byteBlockAllocator, nextTermsHash)}
}

func (f *FreqProxTermsWriter) Flush(ctx context.Context, fieldsToFlush map[string]TermsHashPerField, state *index.SegmentWriteState, sortMap index.DocMap, norms index.NormsProducer) error {

	err := f.BaseTermsHash.Flush(fieldsToFlush, state, sortMap, norms)
	if err != nil {
		return err
	}

	// Gather all fields that saw any postings:
	allFields := make([]*FreqProxTermsWriterPerField, 0)

	for _, field := range fieldsToFlush {
		perField := field.(*FreqProxTermsWriterPerField)
		if perField.getNumTerms() > 0 {
			perField.sortTerms()
			allFields = append(allFields, perField)
		}
	}

	// Sort by field name
	SortFreqProxTermsWriterPerField(allFields)

	fields := NewFreqProxFields(allFields)
	err = f.applyDeletes(state, fields)
	if err != nil {
		return err
	}

	consumer, err := state.SegmentInfo.GetCodec().PostingsFormat().FieldsConsumer(ctx, state)
	if err != nil {
		return err
	}
	defer consumer.Close()
	return consumer.Write(nil, fields, norms)
}

func (f *FreqProxTermsWriter) AddField(invertState *index.FieldInvertState, fieldInfo *document.FieldInfo) (TermsHashPerField, error) {
	addField, err := f.nextTermsHash.AddField(invertState, fieldInfo)
	if err != nil {
		return nil, err
	}
	return NewFreqProxTermsWriterPerField(invertState, f, fieldInfo, addField)
}

func (f *FreqProxTermsWriter) SetTermBytePool(termBytePool *bytesref.BlockPool) {
	f.termBytePool = termBytePool
}

func (f *FreqProxTermsWriter) applyDeletes(state *index.SegmentWriteState, fields index.Fields) error {
	return nil
}
