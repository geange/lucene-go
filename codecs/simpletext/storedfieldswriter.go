package simpletext

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

var (
	FIELDS_EXTENSION         = "fld"
	STORED_FIELD_TYPE_STRING = []byte("string")
	STORED_FIELD_TYPE_BINARY = []byte("binary")
	STORED_FIELD_TYPE_INT    = []byte("int")
	STORED_FIELD_TYPE_LONG   = []byte("long")
	STORED_FIELD_TYPE_FLOAT  = []byte("float")
	STORED_FIELD_TYPE_DOUBLE = []byte("double")
	STORED_FIELD_END         = []byte("END")
	STORED_FIELD_DOC         = []byte("doc ")
	STORED_FIELD_FIELD       = []byte("  field ")
	STORED_FIELD_NAME        = []byte("    name ")
	STORED_FIELD_TYPE        = []byte("    type ")
	STORED_FIELD_VALUE       = []byte("    value ")
)

var _ index.StoredFieldsWriter = &StoredFieldsWriter{}

type StoredFieldsWriter struct {
	numDocsWritten int
	out            store.IndexOutput
	scratch        *bytes.Buffer
}

func newStoredFieldsWriter() *StoredFieldsWriter {
	return &StoredFieldsWriter{}
}

func NewStoredFieldsWriter(ctx context.Context, dir store.Directory, segment string, ioContext *store.IOContext) (*StoredFieldsWriter, error) {
	writer := newStoredFieldsWriter()
	out, err := dir.CreateOutput(ctx, store.SegmentFileName(segment, "", FIELDS_EXTENSION))
	if err != nil {
		return nil, err
	}
	writer.out = out
	return writer, nil
}

func (s *StoredFieldsWriter) Close() error {
	return s.out.Close()
}

func (s *StoredFieldsWriter) StartDocument(context.Context) error {
	if err := s.write(STORED_FIELD_DOC); err != nil {
		return err
	}
	if err := s.write(s.numDocsWritten); err != nil {
		return err
	}
	if err := s.newLine(); err != nil {
		return err
	}
	s.numDocsWritten++
	return nil
}

func (s *StoredFieldsWriter) FinishDocument(context.Context) error {
	return nil
}

func (s *StoredFieldsWriter) WriteField(ctx context.Context, info *document.FieldInfo, field document.IndexableField) error {
	if err := s.write(STORED_FIELD_FIELD); err != nil {
		return err
	}
	if err := s.write(info.Number()); err != nil {
		return err
	}
	if err := s.newLine(); err != nil {
		return err
	}

	if err := s.write(STORED_FIELD_NAME); err != nil {
		return err
	}
	if err := s.write(info.Name()); err != nil {
		return err
	}
	if err := s.newLine(); err != nil {
		return err
	}
	if err := s.write(STORED_FIELD_TYPE); err != nil {
		return err
	}

	obj := field.Get()
	switch v := obj.(type) {
	case int32:
		return s.writeValue(STORED_FIELD_TYPE_INT, fmt.Sprintf("%d", v))
	case int64:
		return s.writeValue(STORED_FIELD_TYPE_LONG, fmt.Sprintf("%d", v))
	case float32:
		value := strconv.FormatFloat(float64(v), 'f', -1, 32)
		return s.writeValue(STORED_FIELD_TYPE_FLOAT, value)
	case float64:
		value := strconv.FormatFloat(v, 'f', -1, 64)
		return s.writeValue(STORED_FIELD_TYPE_DOUBLE, value)
	case string:
		return s.writeValue(STORED_FIELD_TYPE_STRING, v)
	case []byte:
		return s.writeValueBytes(STORED_FIELD_TYPE_BINARY, v)
	default:
		return errors.New("cannot store numeric type")
	}
}

func (s *StoredFieldsWriter) Finish(ctx context.Context, fis *index.FieldInfos, numDocs int) error {
	if s.numDocsWritten != numDocs {
		return errors.New("mergeFields produced an invalid result")
	}
	if err := s.write(STORED_FIELD_END); err != nil {
		return err
	}
	if err := s.newLine(); err != nil {
		return err
	}
	return utils.WriteChecksum(s.out)
}

func (s *StoredFieldsWriter) writeValue(valueType []byte, value string) error {
	if err := utils.WriteBytes(s.out, valueType); err != nil {
		return err
	}

	if err := utils.NewLine(s.out); err != nil {
		return err
	}

	if err := utils.WriteBytes(s.out, STORED_FIELD_VALUE); err != nil {
		return err
	}

	if err := utils.WriteString(s.out, value); err != nil {
		return err
	}

	return utils.NewLine(s.out)
}

func (s *StoredFieldsWriter) writeValueBytes(valueType []byte, value []byte) error {
	if err := utils.WriteBytes(s.out, valueType); err != nil {
		return err
	}

	if err := utils.NewLine(s.out); err != nil {
		return err
	}

	if err := utils.WriteBytes(s.out, STORED_FIELD_VALUE); err != nil {
		return err
	}

	if err := utils.WriteBytes(s.out, value); err != nil {
		return err
	}

	return utils.NewLine(s.out)
}

func (s *StoredFieldsWriter) write(value any) error {
	switch value.(type) {
	case []byte:
		return utils.WriteBytes(s.out, value.([]byte))
	case string:
		return utils.WriteString(s.out, value.(string))
	case int:
		return utils.WriteString(s.out, strconv.Itoa(value.(int)))
	default:
		return nil
	}
}

func (s *StoredFieldsWriter) newLine() error {
	return utils.NewLine(s.out)
}
