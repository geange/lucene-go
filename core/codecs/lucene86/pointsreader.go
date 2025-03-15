package lucene86

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/geange/lucene-go/core/codecs"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util/bkd"
)

var _ index.PointsReader = &PointsReader{}

type PointsReader struct {
	indexIn   store.IndexInput
	dataIn    store.IndexInput
	readState *index.SegmentReadState
	readers   map[int]*bkd.Reader
}

func NewPointsReader(ctx context.Context, readState *index.SegmentReadState) (*PointsReader, error) {
	reader := &PointsReader{
		readState: readState,
		readers:   make(map[int]*bkd.Reader),
	}

	metaFileName := store.SegmentFileName(readState.SegmentInfo.Name(), readState.SegmentSuffix, POINT_META_EXTENSION)
	indexFileName := store.SegmentFileName(readState.SegmentInfo.Name(), readState.SegmentSuffix, POINT_INDEX_EXTENSION)
	dataFileName := store.SegmentFileName(readState.SegmentInfo.Name(), readState.SegmentSuffix, POINT_DATA_EXTENSION)

	indexIn, err := readState.Directory.OpenInput(ctx, indexFileName)
	if err != nil {
		return nil, err
	}
	reader.indexIn = indexIn
	if _, err := codecs.CheckIndexHeader(ctx, indexIn, POINT_INDEX_CODEC_NAME, POINT_VERSION_START,
		POINT_VERSION_CURRENT, readState.SegmentInfo.GetID(), readState.SegmentSuffix); err != nil {
		return nil, err
	}

	dataIn, err := readState.Directory.OpenInput(ctx, dataFileName)
	if err != nil {
		return nil, err
	}
	reader.dataIn = dataIn
	if _, err := codecs.CheckIndexHeader(ctx, dataIn, POINT_DATA_CODEC_NAME, POINT_VERSION_START,
		POINT_VERSION_CURRENT, readState.SegmentInfo.GetID(), readState.SegmentSuffix); err != nil {
		return nil, err
	}

	metaIn, err := store.OpenChecksumInput(ctx, readState.Directory, metaFileName)
	if err != nil {
		return nil, err
	}
	if _, err := codecs.CheckIndexHeader(ctx, metaIn, POINT_META_CODEC_NAME, POINT_VERSION_START, POINT_VERSION_CURRENT,
		readState.SegmentInfo.GetID(), readState.SegmentSuffix); err != nil {
		return nil, err
	}

	for {
		fieldNumber, err := metaIn.ReadUint32(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		bkdReader, err := bkd.NewReader(ctx, metaIn, indexIn, dataIn)
		if err != nil {
			return nil, err
		}
		reader.readers[int(fieldNumber)] = bkdReader
	}
	indexLength, err := metaIn.ReadUint64(ctx)
	if err != nil {
		return nil, err
	}
	dataLength, err := metaIn.ReadUint64(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := codecs.CheckFooter(ctx, metaIn); err != nil {
		return nil, err
	}

	if _, err := codecs.RetrieveChecksumWithLength(ctx, indexIn, int(indexLength)); err != nil {
		return nil, err
	}
	if _, err := codecs.RetrieveChecksumWithLength(ctx, dataIn, int(dataLength)); err != nil {
		return nil, err
	}

	return reader, nil
}

func (p *PointsReader) Close() error {
	if err := p.indexIn.Close(); err != nil {
		return err
	}
	if err := p.dataIn.Close(); err != nil {
		return err
	}
	clear(p.readers)
	return nil
}

func (p *PointsReader) CheckIntegrity() error {
	if _, err := codecs.ChecksumEntireFile(context.Background(), p.indexIn); err != nil {
		return err
	}
	if _, err := codecs.ChecksumEntireFile(context.Background(), p.dataIn); err != nil {
		return err
	}
	return nil
}

func (p *PointsReader) GetValues(ctx context.Context, fieldName string) (types.PointValues, error) {
	fieldInfo := p.readState.FieldInfos.FieldInfo(fieldName)
	if fieldInfo == nil {
		return nil, fmt.Errorf("field=%s is unrecognized", fieldName)
	}
	if fieldInfo.GetPointDimensionCount() == 0 {
		return nil, fmt.Errorf("field=%s did not index point values", fieldName)
	}
	return p.readers[fieldInfo.Number()], nil
}

func (p *PointsReader) GetMergeInstance() index.PointsReader {
	return p
}
