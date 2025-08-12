package perfield

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"maps"
	"slices"
	"strconv"

	"github.com/samber/lo"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
)

const (
	PER_FIELD_NAME       = "PerField40"
	PER_FIELD_FORMAT_KEY = "PerFieldPostingsFormat.format"
	PER_FIELD_SUFFIX_KEY = "PerFieldPostingsFormat.suffix"
)

var _ index.PostingsFormat = &PostingsFormat{}

type PostingsFormat struct {
	getPostingsFormatForField func(field string) index.PostingsFormat
}

func NewPostingsFormat(fn func(field string) index.PostingsFormat) *PostingsFormat {
	return &PostingsFormat{getPostingsFormatForField: fn}
}

func (p *PostingsFormat) GetName() string {
	return PER_FIELD_NAME
}

func (p *PostingsFormat) FieldsConsumer(ctx context.Context, state *index.SegmentWriteState) (index.FieldsConsumer, error) {
	return NewFieldsWriter(state), nil
}

var _ index.FieldsConsumer = &FieldsWriter{}

type FieldsWriter struct {
	format     *PostingsFormat
	writeState *index.SegmentWriteState
	toClose    []io.Closer
}

func NewFieldsWriter(writeState *index.SegmentWriteState) *FieldsWriter {
	return &FieldsWriter{
		writeState: writeState,
		toClose:    make([]io.Closer, 0),
	}
}

func (f *FieldsWriter) Close() error {
	errs := make([]error, 0)
	for _, closer := range f.toClose {
		err := closer.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (f *FieldsWriter) Write(ctx context.Context, fields index.Fields, norms index.NormsProducer) error {
	formatToGroups, err := f.buildFieldsGroupMapping(fields.Iterator())
	if err != nil {
		return err
	}
	for _, entry := range formatToGroups {
		format := entry.K
		group := entry.V

		// Exposes only the fields from this group:
		maskedFields := &filterFields{
			in:    fields,
			group: group,
		}
		consumer, err := format.FieldsConsumer(ctx, group.state)
		if err != nil {
			return err
		}
		f.toClose = append(f.toClose, consumer)
		if err := consumer.Write(ctx, maskedFields, norms); err != nil {
			return err
		}
	}
	return nil
}

var _ index.Fields = &filterFields{}

type filterFields struct {
	in    index.Fields
	group *FieldsGroup
}

func (f *filterFields) Iterator() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, field := range f.group.fields {
			if !yield(field) {
				return
			}
		}
	}
}

func (f *filterFields) Names() []string {
	return f.group.fields
}

func (f *filterFields) Terms(field string) (index.Terms, error) {
	return f.in.Terms(field)
}

func (f *filterFields) Size() int {
	return f.in.Size()
}

func (f *FieldsWriter) buildFieldsGroupMapping(fields iter.Seq[string]) ([]KV[index.PostingsFormat, *FieldsGroup], error) {

	formatToGroupBuilders := make(map[string]*FieldsGroupBuilder)
	formats := make(map[string]index.PostingsFormat)
	suffixes := make(map[string]int)

	for field := range fields {
		fieldInfo := f.writeState.FieldInfos.FieldInfo(field)
		format := f.format.getPostingsFormatForField(field)

		if format == nil {
			return nil, errors.New("invalid null PostingsFormat")
		}
		formatName := format.GetName()

		groupBuilder, ok := formatToGroupBuilders[formatName]
		if !ok {
			suffixes[formatName]++

			suffix := suffixes[formatName]
			segmentSuffix, err := getFullSegmentSuffix(field, f.writeState.SegmentSuffix,
				getSuffix(formatName, fmt.Sprint(suffix)))
			if err != nil {
				return nil, err
			}
			groupBuilder = NewFieldsGroupBuilder(suffix, index.NewSegmentWriteStateWithState(f.writeState, segmentSuffix))
			formatToGroupBuilders[formatName] = groupBuilder
			formats[formatName] = format
		}

		groupBuilder.AddField(field)

		fieldInfo.PutAttribute(PER_FIELD_FORMAT_KEY, formatName)
		fieldInfo.PutAttribute(PER_FIELD_SUFFIX_KEY, strconv.Itoa(groupBuilder.suffix))
	}

	formatToGroups := make([]KV[index.PostingsFormat, *FieldsGroup], 0)
	for formatName, builder := range formatToGroupBuilders {
		item := KV[index.PostingsFormat, *FieldsGroup]{
			K: formats[formatName],
			V: builder.Build(),
		}
		formatToGroups = append(formatToGroups, item)
	}
	return formatToGroups, nil
}

func (p *PostingsFormat) FieldsProducer(ctx context.Context, state *index.SegmentReadState) (index.FieldsProducer, error) {
	return NewFieldsReader(ctx, state)
}

var _ index.FieldsProducer = &FieldsReader{}

type FieldsReader struct {
	segment string
	fields  map[string]index.FieldsProducer
	formats map[string]index.FieldsProducer
}

func NewFieldsReader(ctx context.Context, readState *index.SegmentReadState) (*FieldsReader, error) {
	reader := &FieldsReader{
		segment: readState.SegmentInfo.Name(),
		fields:  make(map[string]index.FieldsProducer),
		formats: make(map[string]index.FieldsProducer),
	}
	// Read field name -> format name
	for _, fi := range readState.FieldInfos.List() {
		if fi.GetIndexOptions() != document.INDEX_OPTIONS_NONE {
			fieldName := fi.Name()
			formatName, ok := fi.GetAttribute(PER_FIELD_FORMAT_KEY)
			if ok {
				// null formatName means the field is in fieldInfos, but has no postings!
				suffix, ok := fi.GetAttribute(PER_FIELD_SUFFIX_KEY)
				if !ok {
					return nil, fmt.Errorf("missing attribute: %s for field: %s", PER_FIELD_SUFFIX_KEY, fieldName)
				}

				format, ok := index.ForNamePostingsFormat(formatName)
				if !ok {
					return nil, fmt.Errorf("format:%s not found", formatName)
				}
				segmentSuffix := getSuffix(formatName, suffix)
				if _, ok := reader.formats[segmentSuffix]; !ok {
					producer, err := format.FieldsProducer(ctx, index.NewSegmentReadStateWithState(readState, segmentSuffix))
					if err != nil {
						return nil, err
					}
					reader.formats[segmentSuffix] = producer
				}
				reader.fields[fieldName] = reader.formats[segmentSuffix]
			}
		}
	}
	return reader, nil
}

func newFieldsReader(other *FieldsReader) *FieldsReader {
	return &FieldsReader{
		segment: other.segment,
		fields:  maps.Clone(other.fields),
		formats: maps.Clone(other.formats),
	}
}

func (f *FieldsReader) Close() error {
	errs := make([]error, 0)
	for _, producer := range f.formats {
		if err := producer.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (f *FieldsReader) Iterator() iter.Seq[string] {
	return func(yield func(string) bool) {
		for k := range f.fields {
			if !yield(k) {
				return
			}
		}
	}
}

func (f *FieldsReader) Names() []string {
	return lo.Keys(f.fields)
}

func (f *FieldsReader) Terms(field string) (index.Terms, error) {
	fieldsProducer, ok := f.fields[field]
	if !ok {
		return nil, errors.New("field not found")
	}
	return fieldsProducer.Terms(field)
}

func (f *FieldsReader) Size() int {
	return len(f.fields)
}

func (f *FieldsReader) CheckIntegrity() error {
	errs := make([]error, 0)
	for _, format := range f.formats {
		if err := format.CheckIntegrity(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (f *FieldsReader) GetMergeInstance() index.FieldsProducer {
	return newFieldsReader(f)
}

type FieldsGroup struct {
	fields []string
	suffix int
	state  *index.SegmentWriteState
}

type FieldsGroupBuilder struct {
	fields map[string]struct{}
	suffix int
	state  *index.SegmentWriteState
}

func NewFieldsGroupBuilder(suffix int, state *index.SegmentWriteState) *FieldsGroupBuilder {
	return &FieldsGroupBuilder{suffix: suffix, state: state, fields: make(map[string]struct{})}
}

func (f *FieldsGroupBuilder) AddField(field string) *FieldsGroupBuilder {
	f.fields[field] = struct{}{}
	return f
}

func (f *FieldsGroupBuilder) Build() *FieldsGroup {
	fields := lo.Keys(f.fields)
	slices.Sort(fields)
	return &FieldsGroup{
		fields: fields,
		suffix: f.suffix,
		state:  f.state,
	}
}

func getSuffix(formatName, suffix string) string {
	return formatName + "_" + suffix
}

func getFullSegmentSuffix(fieldName, outerSegmentSuffix, segmentSuffix string) (string, error) {
	if len(outerSegmentSuffix) == 0 {
		return segmentSuffix, nil
	}
	return "", fmt.Errorf("cannot embed PerFieldPostingsFormat inside itself (field %s returned PerFieldPostingsFormat)", fieldName)
}

type KV[A, B any] struct {
	K A
	V B
}
