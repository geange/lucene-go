package simpletext

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"io"
	"strconv"
)

var _ index.StoredFieldsReader = &SimpleTextStoredFieldsReader{}

// SimpleTextStoredFieldsReader reads plaintext stored fields
// FOR RECREATIONAL USE ONLY
// lucene.experimental
type SimpleTextStoredFieldsReader struct {
	offsets      []int64 // docid -> offset in .fld file
	in           store.IndexInput
	scratch      *bytes.Buffer
	scratchUTF16 *bytes.Buffer
	fieldInfos   *index.FieldInfos
}

func NewSimpleTextStoredFieldsReader(
	directory store.Directory, si *index.SegmentInfo,
	fn *index.FieldInfos, context *store.IOContext) (*SimpleTextStoredFieldsReader, error) {

	fname := store.SegmentFileName(si.Name(), "", FIELDS_EXTENSION)
	input, err := directory.OpenInput(fname, context)
	if err != nil {
		return nil, err
	}

	reader := &SimpleTextStoredFieldsReader{
		offsets:      nil,
		in:           input,
		scratch:      new(bytes.Buffer),
		scratchUTF16: new(bytes.Buffer),
		fieldInfos:   fn,
	}

	maxDoc, err := si.MaxDoc()
	if err != nil {
		return nil, err
	}

	if err := reader.readIndex(maxDoc); err != nil {
		return nil, err
	}
	return reader, nil
}

func newSimpleTextStoredFieldsReader(offsets []int64,
	in store.IndexInput, fieldInfos *index.FieldInfos) *SimpleTextStoredFieldsReader {

	return &SimpleTextStoredFieldsReader{
		offsets:      offsets,
		in:           in,
		fieldInfos:   fieldInfos,
		scratch:      new(bytes.Buffer),
		scratchUTF16: new(bytes.Buffer),
	}
}

func (s *SimpleTextStoredFieldsReader) readIndex(size int) error {
	input := store.NewBufferedChecksumIndexInput(s.in)
	s.offsets = make([]int64, 0, size)
	upto := 0

	reader := utils.NewTextReader(input, s.scratch)

	for !bytes.Equal(s.scratch.Bytes(), STORED_FIELD_END) {
		if err := reader.ReadLine(); err != nil {
			return err
		}

		if bytes.HasPrefix(s.scratch.Bytes(), STORED_FIELD_DOC) {
			s.offsets = append(s.offsets, input.GetFilePointer())
			upto++
		}

	}
	return utils.CheckFooter(input)
}

func (s *SimpleTextStoredFieldsReader) Close() error {
	if err := s.in.Close(); err != nil {
		return err
	}
	s.in = nil
	s.offsets = nil
	return nil
}

func (s *SimpleTextStoredFieldsReader) VisitDocument(docID int, visitor document.StoredFieldVisitor) error {
	if _, err := s.in.Seek(s.offsets[docID], io.SeekStart); err != nil {
		return err
	}

	for {
		if err := s.readLine(); err != nil {
			return err
		}

		if !bytes.HasPrefix(s.scratch.Bytes(), STORED_FIELD_FIELD) {
			break
		}

		fieldNumber, err := parseInt(len(STORED_FIELD_FIELD), s.scratch.Bytes())
		if err != nil {
			return err
		}

		fieldInfo := s.fieldInfos.FieldInfoByNumber(fieldNumber)

		if err := s.readLine(); err != nil {
			return err
		}
		if !bytes.HasPrefix(s.scratch.Bytes(), STORED_FIELD_NAME) {
			return fmt.Errorf("get name error: %s", s.scratch.String())
		}
		if err := s.readLine(); err != nil {
			return err
		}
		if !bytes.HasPrefix(s.scratch.Bytes(), STORED_FIELD_TYPE) {
			return fmt.Errorf("get type error: %s", s.scratch.String())
		}

		lessValues := s.scratch.Bytes()[len(STORED_FIELD_TYPE):]
		var dataType []byte
		switch {
		case bytes.Equal(STORED_FIELD_TYPE_STRING, lessValues):
			dataType = STORED_FIELD_TYPE_STRING
		case bytes.Equal(STORED_FIELD_TYPE_BINARY, lessValues):
			dataType = STORED_FIELD_TYPE_BINARY
		case bytes.Equal(STORED_FIELD_TYPE_INT, lessValues):
			dataType = STORED_FIELD_TYPE_INT
		case bytes.Equal(STORED_FIELD_TYPE_LONG, lessValues):
			dataType = STORED_FIELD_TYPE_LONG
		case bytes.Equal(STORED_FIELD_TYPE_FLOAT, lessValues):
			dataType = STORED_FIELD_TYPE_FLOAT
		case bytes.Equal(STORED_FIELD_TYPE_DOUBLE, lessValues):
			dataType = STORED_FIELD_TYPE_DOUBLE
		default:
			return errors.New("unknown field type")
		}

		status, err := visitor.NeedsField(fieldInfo)
		if err != nil {
			return err
		}

		switch status {
		case document.STORED_FIELD_VISITOR_YES:
			if err := s.readField(dataType, fieldInfo, visitor); err != nil {
				return err
			}
		case document.STORED_FIELD_VISITOR_NO:
			if err := s.readLine(); err != nil {
				return err
			}
			if !bytes.HasPrefix(s.scratch.Bytes(), STORED_FIELD_VALUE) {
				return fmt.Errorf("get value error: %s", s.scratch.String())
			}
		case document.STORED_FIELD_VISITOR_STOP:
			return nil
		default:
			return errors.New("unknown status")
		}
	}
	return nil
}

func (s *SimpleTextStoredFieldsReader) readField(
	dataType []byte, fieldInfo *document.FieldInfo, visitor document.StoredFieldVisitor) error {

	if err := s.readLine(); err != nil {
		return err
	}

	s.scratch.Next(len(STORED_FIELD_VALUE))

	value := s.scratch.Bytes()

	switch {
	case bytes.Equal(STORED_FIELD_TYPE_STRING, dataType):
		return visitor.StringField(fieldInfo, value)
	case bytes.Equal(STORED_FIELD_TYPE_BINARY, dataType):
		return visitor.BinaryField(fieldInfo, value)
	case bytes.Equal(STORED_FIELD_TYPE_INT, dataType):
		num, err := strconv.ParseInt(string(value), 10, 32)
		if err != nil {
			return err
		}
		return visitor.Int32Field(fieldInfo, int32(num))
	case bytes.Equal(STORED_FIELD_TYPE_LONG, dataType):
		num, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return err
		}
		return visitor.Int64Field(fieldInfo, int64(num))
	case bytes.Equal(STORED_FIELD_TYPE_FLOAT, dataType):
		num, err := strconv.ParseFloat(string(value), 32)
		if err != nil {
			return err
		}
		return visitor.Float32Field(fieldInfo, float32(num))
	case bytes.Equal(STORED_FIELD_TYPE_DOUBLE, dataType):
		num, err := strconv.ParseFloat(string(value), 64)
		if err != nil {
			return err
		}
		return visitor.Float64Field(fieldInfo, float64(num))
	default:
		return errors.New("unknown field type")
	}
}

func (s *SimpleTextStoredFieldsReader) readLine() error {
	s.scratch.Reset()
	return utils.ReadLine(s.in, s.scratch)
}

func (s *SimpleTextStoredFieldsReader) Clone() index.StoredFieldsReader {
	if s.in == nil {
		panic("closed!")
	}
	return newSimpleTextStoredFieldsReader(s.offsets, s.in.Clone(), s.fieldInfos)
}

func (s *SimpleTextStoredFieldsReader) CheckIntegrity() error {
	return nil
}

func (s *SimpleTextStoredFieldsReader) GetMergeInstance() index.StoredFieldsReader {
	return s
}

func parseInt(size int, values []byte) (int, error) {
	return strconv.Atoi(string(values[size:]))
}
