package simpletext

import (
	"bytes"
	"errors"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	"strconv"
)

var _ index.SegmentInfoFormat = &SimpleTextSegmentInfoFormat{}

var (
	SI_VERSION     = []byte("    version ")
	SI_MIN_VERSION = []byte("    min version ")
	SI_DOCCOUNT    = []byte("    number of documents ")
	SI_USECOMPOUND = []byte("    uses compound file ")
	SI_NUM_DIAG    = []byte("    diagnostics ")
	SI_DIAG_KEY    = []byte("      key ")
	SI_DIAG_VALUE  = []byte("      value ")
	SI_NUM_ATT     = []byte("    attributes ")
	SI_ATT_KEY     = []byte("      key ")
	SI_ATT_VALUE   = []byte("      value ")
	SI_NUM_FILES   = []byte("    files ")
	SI_FILE        = []byte("      file ")
	SI_ID          = []byte("    id ")
	SI_SORT        = []byte("    sort ")
	SI_SORT_TYPE   = []byte("      type ")
	SI_SORT_NAME   = []byte("      name ")
	SI_SORT_BYTES  = []byte("      bytes ")

	SI_EXTENSION = "si"
)

type SimpleTextSegmentInfoFormat struct {
}

func NewSimpleTextSegmentInfoFormat() *SimpleTextSegmentInfoFormat {
	return &SimpleTextSegmentInfoFormat{}
}

func (s *SimpleTextSegmentInfoFormat) Read(dir store.Directory, segmentName string,
	segmentID []byte, context *store.IOContext) (*index.SegmentInfo, error) {

	scratch := new(bytes.Buffer)
	segFileName := store.SegmentFileName(segmentName, "", SI_EXTENSION)

	input, err := store.OpenChecksumInput(dir, segFileName, context)
	if err != nil {
		return nil, err
	}

	value, err := readValue(input, SI_VERSION, scratch)
	if err != nil {
		return nil, err
	}
	version, err := util.ParseVersion(value)
	if err != nil {
		return nil, err
	}

	var minVersion *util.Version
	value, err = readValue(input, SI_MIN_VERSION, scratch)
	if err != nil {
		return nil, err
	}
	if value == "null" {
		minVersion = nil
	} else {
		minVersion, err = util.ParseVersion(value)
		if err != nil {
			return nil, err
		}
	}

	value, err = readValue(input, SI_DOCCOUNT, scratch)
	if err != nil {
		return nil, err
	}
	docCount, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}

	value, err = readValue(input, SI_USECOMPOUND, scratch)
	if err != nil {
		return nil, err
	}
	isCompoundFile, err := strconv.ParseBool(value)
	if err != nil {
		return nil, err
	}

	value, err = readValue(input, SI_NUM_DIAG, scratch)
	if err != nil {
		return nil, err
	}
	numDiag, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}

	diagnostics := make(map[string]string)

	for i := 0; i < numDiag; i++ {
		key, err := readValue(input, SI_DIAG_KEY, scratch)
		if err != nil {
			return nil, err
		}

		value, err := readValue(input, SI_DIAG_VALUE, scratch)
		if err != nil {
			return nil, err
		}
		diagnostics[key] = value
	}

	value, err = readValue(input, SI_NUM_ATT, scratch)
	if err != nil {
		return nil, err
	}
	numAtt, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}

	attributes := make(map[string]string)

	for i := 0; i < numAtt; i++ {
		key, err := readValue(input, SI_ATT_KEY, scratch)
		if err != nil {
			return nil, err
		}

		value, err := readValue(input, SI_ATT_VALUE, scratch)
		if err != nil {
			return nil, err
		}
		attributes[key] = value
	}

	value, err = readValue(input, SI_NUM_FILES, scratch)
	if err != nil {
		return nil, err
	}
	numFiles, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}

	files := make(map[string]struct{}, numFiles)
	for i := 0; i < numFiles; i++ {
		fileName, err := readValue(input, SI_ATT_KEY, scratch)
		if err != nil {
			return nil, err
		}
		files[fileName] = struct{}{}
	}

	value, err = readValue(input, SI_ID, scratch)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(segmentID, []byte(value)) {
		return nil, errors.New("file mismatch")
	}

	id := make([]byte, len([]byte(value)))
	copy(id, []byte(value))

	value, err = readValue(input, SI_SORT, scratch)
	if err != nil {
		return nil, err
	}
	numSortFields, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	sortField := make([]*index.SortField, 0)
	for i := 0; i < numSortFields; i++ {
		provider, err := readValue(input, SI_SORT_NAME, scratch)
		if err != nil {
			return nil, err
		}

		_, err = readValue(input, SI_SORT_TYPE, scratch)
		if err != nil {
			return nil, err
		}

		value, err = readValue(input, SI_SORT_BYTES, scratch)
		if err != nil {
			return nil, err
		}
		toBytes, err := util.StringToBytes(value)
		if err != nil {
			return nil, err
		}
		output := store.NewByteArrayDataInput(toBytes)
		field, err := index.SingleSortFieldProvider.MustForName(provider).ReadSortField(output)
		if err != nil {
			return nil, err
		}
		sortField = append(sortField, field)
	}

	var indexSort *index.Sort
	if len(sortField) > 0 {
		indexSort = index.NewSort(sortField)
	}

	if err := CheckFooter(input); err != nil {
		return nil, err
	}

	info := index.NewSegmentInfo(dir, version, minVersion, segmentName, docCount,
		isCompoundFile, nil, diagnostics, id, attributes, indexSort)
	info.SetFiles(files)
	return info, nil
}

func (s *SimpleTextSegmentInfoFormat) Write(dir store.Directory,
	si *index.SegmentInfo, ioContext *store.IOContext) error {

	segFileName := store.SegmentFileName(si.Name(), "", SI_EXTENSION)

	output, err := dir.CreateOutput(segFileName, ioContext)
	if err != nil {
		return err
	}

	// Only add the file once we've successfully created it, else IFD assert can trip:
	if err := si.AddFile(segFileName); err != nil {
		return err
	}
	WriteBytes(output, SI_VERSION)
	WriteString(output, si.GetVersion().String())
	WriteNewline(output)

	WriteBytes(output, SI_MIN_VERSION)
	minVersion := si.GetMinVersion()
	if minVersion == nil {
		WriteString(output, "null")
	} else {
		WriteString(output, minVersion.String())
	}
	WriteNewline(output)

	WriteBytes(output, SI_DOCCOUNT)
	maxDoc, _ := si.MaxDoc()
	WriteString(output, strconv.Itoa(maxDoc))
	WriteNewline(output)

	WriteBytes(output, SI_USECOMPOUND)
	WriteString(output, strconv.FormatBool(si.GetUseCompoundFile()))
	WriteNewline(output)

	diagnostics := si.GetDiagnostics()
	numDiagnostics := len(diagnostics)
	WriteBytes(output, SI_NUM_DIAG)
	WriteString(output, strconv.Itoa(numDiagnostics))
	WriteNewline(output)

	for k, v := range diagnostics {
		WriteBytes(output, SI_DIAG_KEY)
		WriteString(output, k)
		WriteNewline(output)

		WriteBytes(output, SI_DIAG_VALUE)
		WriteString(output, v)
		WriteNewline(output)
	}

	attributes := si.GetAttributes()
	WriteBytes(output, SI_NUM_ATT)
	WriteString(output, strconv.Itoa(len(attributes)))
	WriteNewline(output)

	for k, v := range attributes {
		WriteBytes(output, SI_ATT_KEY)
		WriteString(output, k)
		WriteNewline(output)

		WriteBytes(output, SI_ATT_VALUE)
		WriteString(output, v)
		WriteNewline(output)
	}

	files := si.Files()
	WriteBytes(output, SI_NUM_FILES)
	WriteString(output, strconv.Itoa(len(files)))
	WriteNewline(output)

	for fileName := range files {
		WriteBytes(output, SI_FILE)
		WriteString(output, fileName)
		WriteNewline(output)
	}

	WriteBytes(output, SI_ID)
	WriteBytes(output, si.GetID())
	WriteNewline(output)

	indexSort := si.GetIndexSort()
	WriteBytes(output, SI_SORT)
	numSortFields := 0
	if indexSort != nil {
		sortFields := indexSort.GetSort()
		numSortFields = len(sortFields)
	}
	WriteString(output, strconv.Itoa(numSortFields))
	WriteNewline(output)

	if numSortFields > 0 {
		for _, sortField := range indexSort.GetSort() {
			sorter := sortField.GetIndexSorter()
			if sorter == nil {
				return errors.New("cannot serialize sort")
			}

			WriteBytes(output, SI_SORT_NAME)
			WriteString(output, sorter.GetProviderName())
			WriteNewline(output)

			WriteBytes(output, SI_SORT_TYPE)
			WriteString(output, sortField.String())
			WriteNewline(output)

			WriteBytes(output, SI_SORT_BYTES)
			buf := NewBytesOutput()
			index.SingleSortFieldProvider.Write(sortField, buf)
			WriteBytes(output, buf.bytes.Bytes())
			WriteNewline(output)
		}
	}

	return WriteChecksum(output)
}

var _ store.DataOutput = &BytesOutput{}

type BytesOutput struct {
	*store.DataOutputDefault

	bytes *bytes.Buffer
}

func NewBytesOutput() *BytesOutput {
	output := &BytesOutput{bytes: new(bytes.Buffer)}
	output.DataOutputDefault = store.NewDataOutputDefault(&store.DataOutputDefaultConfig{
		WriteByte:  output.WriteByte,
		WriteBytes: output.Write,
	})
	return output
}

func (b *BytesOutput) WriteByte(c byte) error {
	return b.bytes.WriteByte(c)
}

func (b *BytesOutput) Write(bs []byte) (n int, err error) {
	return b.bytes.Write(bs)
}
