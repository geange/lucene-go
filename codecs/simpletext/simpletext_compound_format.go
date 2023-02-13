package simpletext

import (
	"bytes"
	"fmt"
	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"io"
	"math"
	"sort"
	"strconv"
)

const (
	DATA_EXTENSION = "scf"
)

var (
	COMPOUND_FORMAT_HEADER     = []byte("cfs entry for: ")
	COMPOUND_FORMAT_TABLE      = []byte("table of contents, size: ")
	COMPOUND_FORMAT_TABLENAME  = []byte("  filename: ")
	COMPOUND_FORMAT_TABLESTART = []byte("    start: ")
	COMPOUND_FORMAT_TABLEEND   = []byte("    end: ")
	COMPOUND_FORMAT_TABLEPOS   = []byte("table of contents begins at offset: ")
	OFFSETPATTERN              string
)

func init() {
	numDigits := len(strconv.FormatInt(math.MaxInt64, 10))
	pattern := make([]byte, 0, numDigits)
	for i := 0; i < numDigits; i++ {
		pattern = append(pattern, '0')
	}
	OFFSETPATTERN = string(pattern)
}

var _ index.CompoundFormat = &SimpleTextCompoundFormat{}

// SimpleTextCompoundFormat plain text compound format.
// FOR RECREATIONAL USE ONLY
// lucene.experimental
type SimpleTextCompoundFormat struct {
}

func NewSimpleTextCompoundFormat() *SimpleTextCompoundFormat {
	return &SimpleTextCompoundFormat{}
}

func (s *SimpleTextCompoundFormat) Write(dir store.Directory, si *index.SegmentInfo, context *store.IOContext) error {
	dataFile := store.SegmentFileName(si.Name(), "", DATA_EXTENSION)

	numFiles := len(si.Files())
	startOffsets := make([]int64, numFiles)
	endOffsets := make([]int64, numFiles)

	names := make([]string, 0, numFiles)
	for k := range si.Files() {
		names = append(names, k)
	}
	sort.Strings(names)

	out, err := dir.CreateOutput(dataFile, context)
	if err != nil {
		return err
	}

	for i, name := range names {
		// write header for file
		if err := utils.WriteBytes(out, COMPOUND_FORMAT_HEADER); err != nil {
			return err
		}
		if err := utils.WriteString(out, name); err != nil {
			return err
		}
		if err := utils.WriteNewline(out); err != nil {
			return err
		}

		// write bytes for file
		startOffsets[i] = out.GetFilePointer()

		in, err := dir.OpenInput(name, nil)
		if err != nil {
			return err
		}
		if err := out.CopyBytes(in, int(in.Length())); err != nil {
			return err
		}
		endOffsets[i] = out.GetFilePointer()
	}

	tocPos := out.GetFilePointer()

	if err := utils.WriteBytes(out, COMPOUND_FORMAT_TABLE); err != nil {
		return err
	}
	if err := utils.WriteString(out, strconv.Itoa(numFiles)); err != nil {
		return err
	}
	if err := utils.WriteNewline(out); err != nil {
		return err
	}

	for i, name := range names {
		if err := utils.WriteBytes(out, COMPOUND_FORMAT_TABLENAME); err != nil {
			return err
		}
		if err := utils.WriteString(out, name); err != nil {
			return err
		}
		if err := utils.WriteNewline(out); err != nil {
			return err
		}

		if err := utils.WriteBytes(out, COMPOUND_FORMAT_TABLESTART); err != nil {
			return err
		}
		if err := utils.WriteString(out, strconv.Itoa(int(startOffsets[i]))); err != nil {
			return err
		}
		if err := utils.WriteNewline(out); err != nil {
			return err
		}

		if err := utils.WriteBytes(out, COMPOUND_FORMAT_TABLEEND); err != nil {
			return err
		}
		if err := utils.WriteString(out, strconv.Itoa(int(endOffsets[i]))); err != nil {
			return err
		}
		if err := utils.WriteNewline(out); err != nil {
			return err
		}
	}

	if err := utils.WriteBytes(out, COMPOUND_FORMAT_TABLEPOS); err != nil {
		return err
	}
	fmtStr := fmt.Sprintf("%%0%dd", len(OFFSETPATTERN))
	if err := utils.WriteString(out, fmt.Sprintf(fmtStr, tocPos)); err != nil {
		return err
	}
	return utils.WriteNewline(out)
}

func (s *SimpleTextCompoundFormat) GetCompoundReader(dir store.Directory, si *index.SegmentInfo, context *store.IOContext) (index.CompoundDirectory, error) {

	dataFile := store.SegmentFileName(si.Name(), "", DATA_EXTENSION)
	in, err := dir.OpenInput(dataFile, context)
	if err != nil {
		return nil, err
	}

	scratch := new(bytes.Buffer)

	pos := int64(int(in.Length()) - len(COMPOUND_FORMAT_TABLEPOS) - len(OFFSETPATTERN) - 1)
	if _, err := in.Seek(pos, io.SeekStart); err != nil {
		return nil, err
	}

	reader := utils.NewTextReader(in, scratch)
	value, err := reader.ReadLabel(COMPOUND_FORMAT_TABLEPOS)
	if err != nil {
		return nil, err
	}
	tablePos := -1
	tablePos, err = strconv.Atoi(value)
	if err != nil {
		return nil, err
	}

	// seek to TOC and read it
	if _, err := in.Seek(int64(tablePos), io.SeekStart); err != nil {
		return nil, err
	}

	value, err = reader.ReadLabel(COMPOUND_FORMAT_TABLE)
	if err != nil {
		return nil, err
	}
	numEntries, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}

	fileNames := make([]string, 0, numEntries)
	startOffsets := make([]int64, 0, numEntries)
	endOffsets := make([]int64, 0, numEntries)

	for i := 0; i < numEntries; i++ {
		value, err := reader.ReadLabel(COMPOUND_FORMAT_TABLENAME)
		if err != nil {
			return nil, err
		}
		fileNames = append(fileNames, si.Name()+value)

		if i > 0 {
			// files must be unique and in sorted order
			//assert fileNames[i].compareTo(fileNames[i-1]) > 0;
		}

		value, err = reader.ReadLabel(COMPOUND_FORMAT_TABLESTART)
		if err != nil {
			return nil, err
		}
		startOffset, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		startOffsets = append(startOffsets, int64(startOffset))

		value, err = reader.ReadLabel(COMPOUND_FORMAT_TABLESTART)
		if err != nil {
			return nil, err
		}
		endOffset, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		endOffsets = append(endOffsets, int64(endOffset))
	}

	return &innerCompoundDirectory{
		CompoundDirectoryDefault: &index.CompoundDirectoryDefault{},
		in:                       in,
		fileNames:                fileNames,
		startOffsets:             startOffsets,
		endOffsets:               endOffsets,
	}, nil
}

var _ index.CompoundDirectory = &innerCompoundDirectory{}

type innerCompoundDirectory struct {
	*index.CompoundDirectoryDefault

	in           store.IndexInput
	fileNames    []string
	startOffsets []int64
	endOffsets   []int64
}

func (i *innerCompoundDirectory) ListAll() ([]string, error) {
	names := make([]string, len(i.fileNames))
	copy(names, i.fileNames)
	return names, nil
}

func (i *innerCompoundDirectory) FileLength(name string) (int64, error) {
	idx, err := i.getIndex(name)
	if err != nil {
		return 0, err
	}
	return i.endOffsets[idx] - i.startOffsets[idx], nil
}

func (i *innerCompoundDirectory) OpenInput(name string, context *store.IOContext) (store.IndexInput, error) {
	idx, err := i.getIndex(name)
	if err != nil {
		return nil, err
	}
	return i.in.Slice(name, i.startOffsets[idx], i.endOffsets[idx])
}

func (i *innerCompoundDirectory) Close() error {
	return i.in.Close()
}

func (i *innerCompoundDirectory) EnsureOpen() error {
	return nil
}

func (i *innerCompoundDirectory) GetPendingDeletions() (map[string]struct{}, error) {
	return map[string]struct{}{}, nil
}

func (i *innerCompoundDirectory) CheckIntegrity() error {
	return nil
}

func (i *innerCompoundDirectory) getIndex(name string) (int, error) {
	idx := sort.SearchStrings(i.fileNames, name)
	if idx < 0 {
		return 0, fmt.Errorf("no sub-file found: %s", name)
	}
	return idx, nil
}
