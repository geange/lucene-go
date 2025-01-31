package simpletext

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"strconv"

	"github.com/geange/lucene-go/codecs/utils"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/version"
)

var _ index.SegmentInfoFormat = &SegmentInfoFormat{}

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
	SI_EXTENSION   = "si"
)

type SegmentInfoFormat struct {
}

func NewSegmentInfoFormat() *SegmentInfoFormat {
	return &SegmentInfoFormat{}
}

func (s *SegmentInfoFormat) Read(ctx context.Context, dir store.Directory,
	segmentName string, segmentID []byte, ioContext *store.IOContext) (index.SegmentInfo, error) {

	scratch := new(bytes.Buffer)
	segFileName := store.SegmentFileName(segmentName, "", SI_EXTENSION)

	input, err := store.OpenChecksumInput(ctx, dir, segFileName)
	if err != nil {
		return nil, err
	}

	r := utils.NewTextReader(input, scratch)

	value, err := r.ReadLabel(SI_VERSION)
	if err != nil {
		return nil, err
	}
	ver, err := version.Parse(value)
	if err != nil {
		return nil, err
	}

	var minVersion *version.Version
	value, err = r.ReadLabel(SI_MIN_VERSION)
	if err != nil {
		return nil, err
	}
	if value == "null" {
		minVersion = nil
	} else {
		minVersion, err = version.Parse(value)
		if err != nil {
			return nil, err
		}
	}

	docCount, err := r.ParseInt(SI_DOCCOUNT)
	if err != nil {
		return nil, err
	}

	isCompoundFile, err := r.ParseBoolPrefix(SI_USECOMPOUND)
	if err != nil {
		return nil, err
	}

	numDiag, err := r.ParseInt(SI_NUM_DIAG)
	if err != nil {
		return nil, err
	}

	diagnostics := make(map[string]string)

	for i := 0; i < numDiag; i++ {
		diagKey, err := r.ReadLabel(SI_DIAG_KEY)
		if err != nil {
			return nil, err
		}

		diagValue, err := r.ReadLabel(SI_DIAG_VALUE)
		if err != nil {
			return nil, err
		}
		diagnostics[diagKey] = diagValue
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
		attKey, err := r.ReadLabel(SI_ATT_KEY)
		if err != nil {
			return nil, err
		}

		attValue, err := r.ReadLabel(SI_ATT_VALUE)
		if err != nil {
			return nil, err
		}
		attributes[attKey] = attValue
	}

	numFiles, err := r.ParseInt(SI_NUM_FILES)
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

	SIID, err := r.ParseBytes(SI_ID)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(segmentID, SIID) {
		return nil, errors.New("file mismatch")
	}

	id := slices.Clone(SIID)

	numSortFields, err := r.ParseInt(SI_SORT)
	if err != nil {
		return nil, err
	}

	sortField := make([]index.SortField, 0)
	for i := 0; i < numSortFields; i++ {
		provider, err := r.ReadLabel(SI_SORT_NAME)
		if err != nil {
			return nil, err
		}

		if _, err = r.ReadLabel(SI_SORT_TYPE); err != nil {
			return nil, err
		}

		toBytes, err := r.ParseBytes(SI_SORT_BYTES)
		if err != nil {
			return nil, err
		}
		output := store.NewBytesInput(toBytes)
		field, err := coreIndex.GetSortFieldProviderByName(provider).ReadSortField(nil, output)
		if err != nil {
			return nil, err
		}
		sortField = append(sortField, field)
	}

	var indexSort index.Sort
	if len(sortField) > 0 {
		indexSort = coreIndex.NewSort(sortField)
	}

	if err := utils.CheckFooter(input); err != nil {
		return nil, err
	}

	info := coreIndex.NewSegmentInfo(dir, ver, minVersion, segmentName, docCount,
		isCompoundFile, nil, diagnostics, id, attributes, indexSort)
	info.SetFiles(files)
	return info, nil
}

func (s *SegmentInfoFormat) Write(ctx context.Context, dir store.Directory, si index.SegmentInfo, ioContext *store.IOContext) error {

	segFileName := store.SegmentFileName(si.Name(), "", SI_EXTENSION)

	output, err := dir.CreateOutput(ctx, segFileName)
	if err != nil {
		return err
	}

	w := utils.NewTextWriter(output)

	// Only add the file once we've successfully created it, else IFD assert can trip:
	if err := si.AddFile(segFileName); err != nil {
		return err
	}
	if err := w.Bytes(SI_VERSION); err != nil {
		return err
	}
	if err := w.String(si.GetVersion().String()); err != nil {
		return err
	}
	if err := w.NewLine(); err != nil {
		return err
	}

	if err := w.Bytes(SI_MIN_VERSION); err != nil {
		return err
	}
	minVersion := si.GetMinVersion()
	if minVersion == nil {
		if err := w.String("null"); err != nil {
			return err
		}
	} else {
		if err := w.String(minVersion.String()); err != nil {
			return err
		}
	}
	if err := w.NewLine(); err != nil {
		return err
	}

	if err := w.Bytes(SI_DOCCOUNT); err != nil {
		return err
	}
	maxDoc, _ := si.MaxDoc()
	if err := w.String(strconv.Itoa(maxDoc)); err != nil {
		return err
	}
	if err := w.NewLine(); err != nil {
		return err
	}

	if err := w.Bytes(SI_USECOMPOUND); err != nil {
		return err
	}
	if err := w.String(strconv.FormatBool(si.GetUseCompoundFile())); err != nil {
		return err
	}
	if err := w.NewLine(); err != nil {
		return err
	}

	diagnostics := si.GetDiagnostics()
	numDiagnostics := len(diagnostics)
	if err := w.Bytes(SI_NUM_DIAG); err != nil {
		return err
	}
	if err := w.String(strconv.Itoa(numDiagnostics)); err != nil {
		return err
	}
	if err := w.NewLine(); err != nil {
		return err
	}

	for k, v := range diagnostics {
		if err := w.Bytes(SI_DIAG_KEY); err != nil {
			return err
		}
		if err := w.String(k); err != nil {
			return err
		}
		if err := w.NewLine(); err != nil {
			return err
		}

		if err := w.Bytes(SI_DIAG_VALUE); err != nil {
			return err
		}
		if err := w.String(v); err != nil {
			return err
		}
		if err := w.NewLine(); err != nil {
			return err
		}
	}

	attributes := si.GetAttributes()
	if err := w.Bytes(SI_NUM_ATT); err != nil {
		return err
	}
	if err := w.String(strconv.Itoa(len(attributes))); err != nil {
		return err
	}
	if err := w.NewLine(); err != nil {
		return err
	}

	for k, v := range attributes {
		if err := w.Bytes(SI_ATT_KEY); err != nil {
			return err
		}
		if err := w.String(k); err != nil {
			return err
		}
		if err := w.NewLine(); err != nil {
			return err
		}

		if err := w.Bytes(SI_ATT_VALUE); err != nil {
			return err
		}
		if err := w.String(v); err != nil {
			return err
		}
		if err := w.NewLine(); err != nil {
			return err
		}
	}

	if err := w.Bytes(SI_NUM_FILES); err != nil {
		return err
	}
	if err := w.String(strconv.Itoa(si.FilesNum())); err != nil {
		return err
	}
	if err := w.NewLine(); err != nil {
		return err
	}

	for fName := range si.Files() {
		if err := w.Bytes(SI_FILE); err != nil {
			return err
		}
		if err := w.String(fName); err != nil {
			return err
		}
		if err := w.NewLine(); err != nil {
			return err
		}
	}

	if err := w.Bytes(SI_ID); err != nil {
		return err
	}
	if err := w.Bytes(si.GetID()); err != nil {
		return err
	}
	if err := w.NewLine(); err != nil {
		return err
	}

	indexSort := si.GetIndexSort()
	if err := w.Bytes(SI_SORT); err != nil {
		return err
	}
	numSortFields := 0
	if indexSort != nil {
		sortFields := indexSort.GetSort()
		numSortFields = len(sortFields)
	}
	if err := w.String(strconv.Itoa(numSortFields)); err != nil {
		return err
	}
	if err := w.NewLine(); err != nil {
		return err
	}

	if numSortFields > 0 {
		for _, sortField := range indexSort.GetSort() {
			sorter := sortField.GetIndexSorter()
			if sorter == nil {
				return errors.New("cannot serialize sort")
			}

			if err := w.Bytes(SI_SORT_NAME); err != nil {
				return err
			}
			if err := w.String(sorter.GetProviderName()); err != nil {
				return err
			}
			if err := w.NewLine(); err != nil {
				return err
			}

			if err := w.Bytes(SI_SORT_TYPE); err != nil {
				return err
			}
			if err := w.String(sortField.String()); err != nil {
				return err
			}
			if err := w.NewLine(); err != nil {
				return err
			}

			if err := w.Bytes(SI_SORT_BYTES); err != nil {
				return err
			}
			buf := NewBytesOutput()
			if err := coreIndex.WriteSortField(sortField, buf); err != nil {
				return err
			}
			if err := w.Bytes(buf.bytes.Bytes()); err != nil {
				return err
			}
			if err := w.NewLine(); err != nil {
				return err
			}
		}
	}

	if err := utils.WriteChecksum(output); err != nil {
		return err
	}
	return output.Close()
}

var _ store.DataOutput = &BytesOutput{}

type BytesOutput struct {
	*store.BaseDataOutput

	bytes *bytes.Buffer
}

func NewBytesOutput() *BytesOutput {
	output := &BytesOutput{bytes: new(bytes.Buffer)}
	output.BaseDataOutput = store.NewBaseDataOutput(output)
	return output
}

func (b *BytesOutput) WriteByte(c byte) error {
	return b.bytes.WriteByte(c)
}

func (b *BytesOutput) Write(bs []byte) (n int, err error) {
	return b.bytes.Write(bs)
}
