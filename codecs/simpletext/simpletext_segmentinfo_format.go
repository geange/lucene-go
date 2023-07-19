package simpletext

import (
	"bytes"
	"errors"
	"github.com/geange/lucene-go/codecs/utils"
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

	r := utils.NewTextReader(input, scratch)

	value, err := r.ReadLabel(SI_VERSION)
	if err != nil {
		return nil, err
	}
	version, err := util.ParseVersion(value)
	if err != nil {
		return nil, err
	}

	var minVersion *util.Version
	value, err = r.ReadLabel(SI_MIN_VERSION)
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

	value, err = r.ReadLabel(SI_DOCCOUNT)
	if err != nil {
		return nil, err
	}
	docCount, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}

	value, err = r.ReadLabel(SI_USECOMPOUND)
	if err != nil {
		return nil, err
	}
	isCompoundFile, err := strconv.ParseBool(value)
	if err != nil {
		return nil, err
	}

	value, err = r.ReadLabel(SI_NUM_DIAG)
	if err != nil {
		return nil, err
	}
	numDiag, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}

	diagnostics := make(map[string]string)

	for i := 0; i < numDiag; i++ {
		key, err := r.ReadLabel(SI_DIAG_KEY)
		if err != nil {
			return nil, err
		}

		value, err := r.ReadLabel(SI_DIAG_VALUE)
		if err != nil {
			return nil, err
		}
		diagnostics[key] = value
	}

	value, err = r.ReadLabel(SI_NUM_ATT)
	if err != nil {
		return nil, err
	}
	numAtt, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}

	attributes := make(map[string]string)

	for i := 0; i < numAtt; i++ {
		key, err := r.ReadLabel(SI_ATT_KEY)
		if err != nil {
			return nil, err
		}

		value, err := r.ReadLabel(SI_ATT_VALUE)
		if err != nil {
			return nil, err
		}
		attributes[key] = value
	}

	value, err = r.ReadLabel(SI_NUM_FILES)
	if err != nil {
		return nil, err
	}
	numFiles, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}

	files := make(map[string]struct{}, numFiles)
	for i := 0; i < numFiles; i++ {
		fileName, err := r.ReadLabel(SI_FILE)
		if err != nil {
			return nil, err
		}
		files[fileName] = struct{}{}
	}

	value, err = r.ReadLabel(SI_ID)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(segmentID, []byte(value)) {
		return nil, errors.New("file mismatch")
	}

	id := make([]byte, len([]byte(value)))
	copy(id, []byte(value))

	value, err = r.ReadLabel(SI_SORT)
	if err != nil {
		return nil, err
	}
	numSortFields, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	sortField := make([]index.SortField, 0)
	for i := 0; i < numSortFields; i++ {
		provider, err := r.ReadLabel(SI_SORT_NAME)
		if err != nil {
			return nil, err
		}

		_, err = r.ReadLabel(SI_SORT_TYPE)
		if err != nil {
			return nil, err
		}

		value, err = r.ReadLabel(SI_SORT_BYTES)
		if err != nil {
			return nil, err
		}
		toBytes, err := util.StringToBytes(value)
		if err != nil {
			return nil, err
		}
		output := store.NewByteArrayDataInput(toBytes)
		field, err := index.GetSortFieldProviderByName(provider).ReadSortField(output)
		if err != nil {
			return nil, err
		}
		sortField = append(sortField, field)
	}

	var indexSort *index.Sort
	if len(sortField) > 0 {
		indexSort = index.NewSort(sortField)
	}

	if err := utils.CheckFooter(input); err != nil {
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

	w := utils.NewTextWriter(output)

	// Only add the file once we've successfully created it, else IFD assert can trip:
	if err := si.AddFile(segFileName); err != nil {
		return err
	}
	w.Bytes(SI_VERSION)
	w.String(si.GetVersion().String())
	w.NewLine()

	w.Bytes(SI_MIN_VERSION)
	minVersion := si.GetMinVersion()
	if minVersion == nil {
		w.String("null")
	} else {
		w.String(minVersion.String())
	}
	w.NewLine()

	w.Bytes(SI_DOCCOUNT)
	maxDoc, _ := si.MaxDoc()
	w.String(strconv.Itoa(maxDoc))
	w.NewLine()

	w.Bytes(SI_USECOMPOUND)
	w.String(strconv.FormatBool(si.GetUseCompoundFile()))
	w.NewLine()

	diagnostics := si.GetDiagnostics()
	numDiagnostics := len(diagnostics)
	w.Bytes(SI_NUM_DIAG)
	w.String(strconv.Itoa(numDiagnostics))
	w.NewLine()

	for k, v := range diagnostics {
		w.Bytes(SI_DIAG_KEY)
		w.String(k)
		w.NewLine()

		w.Bytes(SI_DIAG_VALUE)
		w.String(v)
		w.NewLine()
	}

	attributes := si.GetAttributes()
	w.Bytes(SI_NUM_ATT)
	w.String(strconv.Itoa(len(attributes)))
	w.NewLine()

	for k, v := range attributes {
		w.Bytes(SI_ATT_KEY)
		w.String(k)
		w.NewLine()

		w.Bytes(SI_ATT_VALUE)
		w.String(v)
		w.NewLine()
	}

	files := si.Files()
	w.Bytes(SI_NUM_FILES)
	w.String(strconv.Itoa(len(files)))
	w.NewLine()

	for fileName := range files {
		w.Bytes(SI_FILE)
		w.String(fileName)
		w.NewLine()
	}

	w.Bytes(SI_ID)
	w.Bytes(si.GetID())
	w.NewLine()

	indexSort := si.GetIndexSort()
	w.Bytes(SI_SORT)
	numSortFields := 0
	if indexSort != nil {
		sortFields := indexSort.GetSort()
		numSortFields = len(sortFields)
	}
	w.String(strconv.Itoa(numSortFields))
	w.NewLine()

	if numSortFields > 0 {
		for _, sortField := range indexSort.GetSort() {
			sorter := sortField.GetIndexSorter()
			if sorter == nil {
				return errors.New("cannot serialize sort")
			}

			w.Bytes(SI_SORT_NAME)
			w.String(sorter.GetProviderName())
			w.NewLine()

			w.Bytes(SI_SORT_TYPE)
			w.String(sortField.String())
			w.NewLine()

			w.Bytes(SI_SORT_BYTES)
			buf := NewBytesOutput()
			index.WriteSortField(sortField, buf)
			w.Bytes(buf.bytes.Bytes())
			w.NewLine()
		}
	}

	return utils.WriteChecksum(output)
}

var _ store.DataOutput = &BytesOutput{}

type BytesOutput struct {
	*store.Writer

	bytes *bytes.Buffer
}

func NewBytesOutput() *BytesOutput {
	output := &BytesOutput{bytes: new(bytes.Buffer)}
	output.Writer = store.NewWriter(output)
	return output
}

func (b *BytesOutput) WriteByte(c byte) error {
	return b.bytes.WriteByte(c)
}

func (b *BytesOutput) Write(bs []byte) (n int, err error) {
	return b.bytes.Write(bs)
}
