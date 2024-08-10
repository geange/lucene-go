package index

import (
	"context"
	"fmt"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	"strconv"
)

var _ index.CodecReader = &SegmentReader{}

// SegmentReader
// IndexReader implementation over a single segment.
// Instances pointing to the same segment (but with different deletes, etc) may share the same core data.
// lucene.experimental
type SegmentReader struct {
	*BaseCodecReader

	si index.SegmentCommitInfo

	// this is the original SI that IW uses internally but it's mutated behind the scenes
	// and we don't want this SI to be used for anything. Yet, IW needs this to do maintainance
	// and lookup pooled readers etc.
	originalSi index.SegmentCommitInfo

	metaData     index.LeafMetaData
	liveDocs     util.Bits
	hardLiveDocs util.Bits

	// Normally set to si.maxDoc - si.delDocCount, unless we
	// were created as an NRT reader from IW, in which case IW
	// tells us the number of live docs:
	numDocs int

	core         *SegmentCoreReaders
	segDocValues *SegmentDocValues

	// True if we are holding RAM only liveDocs or DV updates,
	// i.e. the SegmentCommitInfo delGen doesn't match our liveDocs.
	isNRT bool

	docValuesProducer index.DocValuesProducer

	fieldInfos index.FieldInfos
}

// NewSegmentReader
// Constructs a new SegmentReader with a new core.
func NewSegmentReader(ctx context.Context, si index.SegmentCommitInfo,
	createdVersionMajor int, ioContext *store.IOContext) (*SegmentReader, error) {

	readers, err := NewSegmentCoreReaders(ctx, si.Info().Dir(), si, ioContext)
	if err != nil {
		return nil, err
	}

	reader := &SegmentReader{
		si:                si.Clone(),
		originalSi:        si,
		metaData:          NewLeafMetaData(createdVersionMajor, si.Info().GetMinVersion(), si.Info().GetIndexSort()),
		liveDocs:          nil,
		hardLiveDocs:      nil,
		numDocs:           0,
		core:              readers,
		segDocValues:      NewSegmentDocValues(),
		isNRT:             false, // We pull liveDocs/DV updates from disk:
		docValuesProducer: nil,
		fieldInfos:        nil,
	}

	reader.BaseCodecReader = NewBaseCodecReader(reader)

	codec := si.Info().GetCodec()
	if si.HasDeletions() {
		// NOTE: the bitvector is stored using the regular directory, not cfs
		liveDocs, err := codec.LiveDocsFormat().ReadLiveDocs(ctx, reader.Directory(), si, ioContext)
		if err != nil {
			return nil, err
		}
		reader.hardLiveDocs = liveDocs
		reader.liveDocs = liveDocs
	} else {
		reader.hardLiveDocs = nil
		reader.liveDocs = nil
	}

	maxDoc, err := si.Info().MaxDoc()
	if err != nil {
		return nil, err
	}
	reader.numDocs = maxDoc - si.GetDelCount()

	fieldInfos, err := reader.initFieldInfos()
	if err != nil {
		return nil, err
	}
	reader.fieldInfos = fieldInfos

	docValuesProducer, err := reader.initDocValuesProducer()
	if err != nil {
		return nil, err
	}
	reader.docValuesProducer = docValuesProducer
	return reader, nil
}

// New
// Create new SegmentReader sharing core from a previous SegmentReader and using the provided liveDocs,
// and recording whether those liveDocs were carried in ram (isNRT=true).
func (s *SegmentReader) New(si index.SegmentCommitInfo, liveDocs, hardLiveDocs util.Bits, numDocs int, isNRT bool) (*SegmentReader, error) {

	maxDoc, err := si.Info().MaxDoc()
	if err != nil {
		return nil, err
	}
	if numDocs > maxDoc {
		return nil, fmt.Errorf("numDocs=%d but maxDoc=%d", numDocs, maxDoc)
	}

	if liveDocs != nil && int(liveDocs.Len()) != maxDoc {
		return nil, fmt.Errorf("maxDoc=%d but liveDocs.size()=%d", maxDoc, liveDocs.Len())
	}

	reader := &SegmentReader{
		si:                si.Clone(),
		originalSi:        si,
		metaData:          s.GetMetaData(),
		liveDocs:          liveDocs,
		hardLiveDocs:      hardLiveDocs,
		numDocs:           numDocs,
		core:              s.core,
		segDocValues:      s.segDocValues,
		isNRT:             isNRT,
		docValuesProducer: nil,
		fieldInfos:        nil,
	}

	if err := reader.core.incRef(); err != nil {
		return nil, err
	}

	fieldInfos, err := reader.initFieldInfos()
	if err != nil {
		if err := reader.DoClose(); err != nil {
			return nil, err
		}
		return nil, err
	}
	reader.fieldInfos = fieldInfos

	docValuesProducer, err := reader.initDocValuesProducer()
	if err != nil {
		if err := reader.DoClose(); err != nil {
			return nil, err
		}
		return nil, err
	}
	reader.docValuesProducer = docValuesProducer

	return reader, nil
}

func (s *SegmentReader) Directory() store.Directory {
	return s.si.Info().Dir()
}

func (s *SegmentReader) NumDocs() int {
	// Don't call ensureOpen() here (it could affect performance)
	return s.numDocs
}

func (s *SegmentReader) MaxDoc() int {
	// Don't call ensureOpen() here (it could affect performance)
	v, _ := s.si.Info().MaxDoc()
	return v
}

func (s *SegmentReader) DoClose() error {
	if err := s.core.decRef(); err != nil {
		return err
	}

	if producer, ok := s.docValuesProducer.(*SegmentDocValuesProducer); ok {
		if err := s.segDocValues.decRef(producer.dvGens); err != nil {
			return err
		} else if s.docValuesProducer != nil {
			return s.segDocValues.decRef([]int64{-1})
		}
	}
	return nil
}

func (s *SegmentReader) GetReaderCacheHelper() index.CacheHelper {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentReader) GetFieldInfos() index.FieldInfos {
	return s.fieldInfos
}

func (s *SegmentReader) GetLiveDocs() util.Bits {
	return s.liveDocs
}

func (s *SegmentReader) GetMetaData() index.LeafMetaData {
	return s.metaData
}

func (s *SegmentReader) GetFieldsReader() index.StoredFieldsReader {
	return s.core.fieldsReaderLocal
}

func (s *SegmentReader) GetTermVectorsReader() index.TermVectorsReader {
	//ensureOpen();
	return s.core.termVectorsLocal
}

func (s *SegmentReader) GetNormsReader() index.NormsProducer {
	return s.core.normsProducer
}

func (s *SegmentReader) GetDocValuesReader() index.DocValuesProducer {
	//ensureOpen();
	return s.docValuesProducer
}

func (s *SegmentReader) GetPostingsReader() index.FieldsProducer {
	//ensureOpen();
	return s.core.fields
}

func (s *SegmentReader) GetPointsReader() index.PointsReader {
	return s.core.pointsReader
}

// GetOriginalSegmentInfo
// Returns the original SegmentInfo passed to the segment reader on creation time.
// getSegmentInfo() returns a clone of this instance.
func (s *SegmentReader) GetOriginalSegmentInfo() index.SegmentCommitInfo {
	return s.originalSi
}

// GetHardLiveDocs
// Returns the live docs that are not hard-deleted. This is an expert API to be used with soft-deletes
// to filter out document that hard deleted for instance due to aborted documents or to distinguish
// soft and hard deleted documents ie. a rolled back tombstone.
// lucene.experimental
func (s *SegmentReader) GetHardLiveDocs() util.Bits {
	return s.hardLiveDocs
}

func (s *SegmentReader) initFieldInfos() (index.FieldInfos, error) {
	if !s.si.HasFieldUpdates() {
		return s.core.coreFieldInfos, nil
	} else {
		// updates always outside of CFS
		fisFormat := s.si.Info().GetCodec().FieldInfosFormat()
		segmentSuffix := strconv.FormatInt(s.si.GetFieldInfosGen(), 36)
		return fisFormat.Read(nil, s.si.Info().Dir(), s.si.Info(), segmentSuffix, nil)
	}
}

// init most recent DocValues for the current commit
func (s *SegmentReader) initDocValuesProducer() (index.DocValuesProducer, error) {
	if s.fieldInfos.HasDocValues() == false {
		return nil, nil
	} else {
		var dir store.Directory
		if s.core.cfsReader != nil {
			dir = s.core.cfsReader
		} else {
			dir = s.si.Info().Dir()
		}
		if s.si.HasFieldUpdates() {
			return NewSegmentDocValuesProducer(s.si, dir, s.core.coreFieldInfos, s.fieldInfos, s.segDocValues)
		} else {
			// simple case, no DocValues updates
			return s.segDocValues.GetDocValuesProducer(-1, s.si, dir, s.fieldInfos)
		}
	}
}
