package lucene86

import (
	"context"
	"errors"
	"fmt"

	"github.com/geange/lucene-go/core/codecs"
	index2 "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/version"
)

var _ index.SegmentInfoFormat = &SegmentInfoFormat{}

type SegmentInfoFormat struct {
}

func NewSegmentInfoFormat() *SegmentInfoFormat {
	return &SegmentInfoFormat{}
}

func (s *SegmentInfoFormat) Read(ctx context.Context, dir store.Directory, segment string, segmentID []byte, ioContext *store.IOContext) (index.SegmentInfo, error) {
	fileName := store.SegmentFileName(segment, "", SI_EXTENSION)
	input, err := store.OpenChecksumInput(ctx, dir, fileName)
	if err != nil {
		return nil, err
	}
	if _, err := codecs.CheckIndexHeader(ctx, input, SI_CODEC_NAME, SI_VERSION_START,
		SI_VERSION_CURRENT, segmentID, ""); err != nil {
		return nil, err
	}

	major, err := input.ReadUint32(ctx)
	if err != nil {
		return nil, err
	}
	minor, err := input.ReadUint32(ctx)
	if err != nil {
		return nil, err
	}
	bugfix, err := input.ReadUint32(ctx)
	if err != nil {
		return nil, err
	}
	rVersion, err := version.New(
		version.WithMajor(uint8(major)),
		version.WithMinor(uint8(minor)),
		version.WithBugfix(uint8(bugfix)))
	hasMinVersion, err := input.ReadByte()
	if err != nil {
		return nil, err
	}
	var minVersion *version.Version
	if hasMinVersion == 1 {
		minMajor, err := input.ReadUint32(ctx)
		if err != nil {
			return nil, err
		}
		minMinor, err := input.ReadUint32(ctx)
		if err != nil {
			return nil, err
		}
		minBugfix, err := input.ReadUint32(ctx)
		if err != nil {
			return nil, err
		}
		mVersion, err := version.New(
			version.WithMajor(uint8(minMajor)),
			version.WithMinor(uint8(minMinor)),
			version.WithBugfix(uint8(minBugfix)))
		if err != nil {
			return nil, err
		}
		minVersion = mVersion
	}

	docCount, err := input.ReadUint32(ctx)
	if int32(docCount) < 0 {
		return nil, fmt.Errorf("invalid docCount: %d", int32(docCount))
	}

	isCompoundFileFlag, err := input.ReadByte()
	if err != nil {
		return nil, err
	}
	isCompoundFile := isCompoundFileFlag == byte(index2.SegmentInfoYES)

	diagnostics, err := input.ReadMapOfStrings(ctx)
	if err != nil {
		return nil, err
	}
	files, err := input.ReadSetOfStrings(ctx)
	if err != nil {
		return nil, err
	}
	attributes, err := input.ReadMapOfStrings(ctx)
	if err != nil {
		return nil, err
	}

	numSortFields, err := input.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}
	var indexSort index.Sort
	if numSortFields > 0 {

		sortFields := make([]index.SortField, numSortFields)
		for i := 0; i < int(numSortFields); i++ {
			name, err := input.ReadString(ctx)
			if err != nil {
				return nil, err
			}
			sortField, err := index2.GetSortFieldProviderByName(name).ReadSortField(ctx, input)
			if err != nil {
				return nil, err
			}

			sortFields[i] = sortField
		}
		indexSort = index2.NewSort(sortFields)
	} else if numSortFields < 0 {
		return nil, fmt.Errorf("invalid index sort field count: %d", numSortFields)
	}

	si := index2.NewSegmentInfo(dir, rVersion, minVersion, segment, int(docCount), isCompoundFile,
		nil, diagnostics, segmentID, attributes, indexSort)
	si.SetFiles(files)

	return si, nil
}

func (s *SegmentInfoFormat) Write(ctx context.Context, dir store.Directory, si index.SegmentInfo, ioContext *store.IOContext) error {
	fileName := store.SegmentFileName(si.Name(), "", SI_EXTENSION)
	output, err := dir.CreateOutput(ctx, fileName)
	if err != nil {
		return err
	}
	// Only add the file once we've successfully created it, else IFD assert can trip:
	if err := si.AddFile(fileName); err != nil {
		return err
	}
	if err := codecs.WriteIndexHeader(ctx, output, SI_CODEC_NAME, SI_VERSION_CURRENT,
		si.GetID(), ""); err != nil {
		return err
	}
	segVersion := si.GetVersion()
	if segVersion.Major() < 7 {
		return fmt.Errorf("invalid major version: should be >= 7 but got: %d", segVersion.Major())
	}

	// Write the Lucene version that created this segment, since 3.1
	if err := output.WriteUint32(ctx, uint32(segVersion.Major())); err != nil {
		return err
	}
	if err := output.WriteUint32(ctx, uint32(segVersion.Minor())); err != nil {
		return err
	}
	if err := output.WriteUint32(ctx, uint32(segVersion.Bugfix())); err != nil {
		return err
	}

	// Write the min Lucene version that contributed docs to the segment, since 7.0
	if si.GetMinVersion() != nil {
		if err := output.WriteByte(1); err != nil {
			return err
		}
		minVersion := si.GetMinVersion()
		if err := output.WriteUint32(ctx, uint32(minVersion.Major())); err != nil {
			return err
		}
		if err := output.WriteUint32(ctx, uint32(minVersion.Minor())); err != nil {
			return err
		}
		if err := output.WriteUint32(ctx, uint32(minVersion.Bugfix())); err != nil {
			return err
		}
	} else {
		if err := output.WriteByte(0); err != nil {
			return err
		}
	}
	maxDoc, err := si.MaxDoc()
	if err != nil {
		return err
	}
	if err := output.WriteUint32(ctx, uint32(maxDoc)); err != nil {
		return err
	}

	isUseCompoundFile := index2.SegmentInfoNO
	if si.GetUseCompoundFile() {
		isUseCompoundFile = index2.SegmentInfoYES
	}
	if err := output.WriteByte(byte(isUseCompoundFile)); err != nil {
		return err
	}
	if err := output.WriteMapOfStrings(ctx, si.GetDiagnostics()); err != nil {
		return err
	}
	files := si.Files()
	for file := range files {
		if index2.ParseSegmentName(file) != si.Name() {
			return fmt.Errorf("invalid files: expected segment=%s, got=%s", si.Name(), file)
		}
	}
	if err := output.WriteSetOfStrings(ctx, files); err != nil {
		return err
	}
	if err := output.WriteMapOfStrings(ctx, si.GetAttributes()); err != nil {
		return err
	}
	indexSort := si.GetIndexSort()

	sortFields := indexSort.GetSort()
	if err := output.WriteUvarint(ctx, uint64(len(sortFields))); err != nil {
		return err
	}
	for _, sortField := range indexSort.GetSort() {
		sorter := sortField.GetIndexSorter()
		if sorter == nil {
			return errors.New("cannot serialize SortField")
		}
		if err := output.WriteString(ctx, sorter.GetProviderName()); err != nil {
			return err
		}

		if err := index2.SortFieldProviderWrite(ctx, sortField, output); err != nil {
			return err
		}
	}

	if err := codecs.WriteFooter(ctx, output); err != nil {
		return err
	}

	return nil
}

const (
	SI_EXTENSION       = "si"
	SI_CODEC_NAME      = "Lucene86SegmentInfo"
	SI_VERSION_START   = 0
	SI_VERSION_CURRENT = SI_VERSION_START
)
