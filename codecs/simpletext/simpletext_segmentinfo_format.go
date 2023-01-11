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
		field, err := index.LooksUpSortFieldProviderByName(provider).ReadSortField(output)
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
	info *index.SegmentInfo, ioContext *store.IOContext) error {
	//TODO implement me
	panic("implement me")
}
