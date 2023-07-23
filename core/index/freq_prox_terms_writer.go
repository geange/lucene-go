package index

import (
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/util"
)

var _ TermsHash = &FreqProxTermsWriter{}

type FreqProxTermsWriter struct {
	*TermsHashDefault
}

func NewFreqProxTermsWriter(intBlockAllocator util.IntsAllocator,
	byteBlockAllocator util.BytesAllocator, nextTermsHash TermsHash) *FreqProxTermsWriter {

	return &FreqProxTermsWriter{
		NewTermsHashDefault(intBlockAllocator, byteBlockAllocator, nextTermsHash)}
}

func (f *FreqProxTermsWriter) Flush(fieldsToFlush map[string]TermsHashPerField,
	state *SegmentWriteState, sortMap *DocMap, norms NormsProducer) error {

	err := f.TermsHashDefault.Flush(fieldsToFlush, state, sortMap, norms)
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

	consumer, err := state.SegmentInfo.GetCodec().PostingsFormat().FieldsConsumer(state)
	if err != nil {
		return err
	}
	defer consumer.Close()
	return consumer.Write(fields, norms)
}

func (f *FreqProxTermsWriter) AddField(invertState *FieldInvertState, fieldInfo *document.FieldInfo) (TermsHashPerField, error) {
	addField, err := f.nextTermsHash.AddField(invertState, fieldInfo)
	if err != nil {
		return nil, err
	}
	return NewFreqProxTermsWriterPerField(invertState, f, fieldInfo, addField), nil
}

func (f *FreqProxTermsWriter) SetTermBytePool(termBytePool *util.ByteBlockPool) {
	f.termBytePool = termBytePool
}

func (f *FreqProxTermsWriter) applyDeletes(state *SegmentWriteState, fields Fields) error {
	return nil
}
