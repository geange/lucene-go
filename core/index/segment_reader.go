package index

import (
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	"strconv"
)

var _ CodecReader = &SegmentReader{}

// SegmentReader
// IndexReader implementation over a single segment.
// Instances pointing to the same segment (but with different deletes, etc) may share the same core data.
// lucene.experimental
type SegmentReader struct {
	*CodecReaderDefault

	si *SegmentCommitInfo

	// this is the original SI that IW uses internally but it's mutated behind the scenes
	// and we don't want this SI to be used for anything. Yet, IW needs this to do maintainance
	// and lookup pooled readers etc.
	originalSi *SegmentCommitInfo

	metaData *LeafMetaData

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

	docValuesProducer DocValuesProducer

	fieldInfos *FieldInfos
}

func NewSegmentReader(si *SegmentCommitInfo,
	createdVersionMajor int, context *store.IOContext) (*SegmentReader, error) {

	readers, err := NewSegmentCoreReaders(si.info.dir, si, context)
	if err != nil {
		return nil, err
	}

	reader := &SegmentReader{
		si:                si.Clone(),
		originalSi:        si,
		metaData:          NewLeafMetaData(createdVersionMajor, si.info.GetMinVersion(), si.info.GetIndexSort()),
		liveDocs:          nil,
		hardLiveDocs:      nil,
		numDocs:           0,
		core:              readers,
		segDocValues:      NewSegmentDocValues(),
		isNRT:             false, // We pull liveDocs/DV updates from disk:
		docValuesProducer: nil,
		fieldInfos:        nil,
	}

	reader.CodecReaderDefault = NewCodecReaderDefault(reader)

	codec := si.info.GetCodec()
	if si.HasDeletions() {
		// NOTE: the bitvector is stored using the regular directory, not cfs
		liveDocs, err := codec.LiveDocsFormat().ReadLiveDocs(reader.Directory(), si, nil)
		if err != nil {
			return nil, err
		}
		reader.hardLiveDocs = liveDocs
		reader.liveDocs = liveDocs
	} else {
		//assert si.getDelCount() == 0;
		reader.hardLiveDocs = nil
		reader.liveDocs = nil
	}

	maxDoc, err := si.info.MaxDoc()
	if err != nil {
		return nil, err
	}
	reader.numDocs = maxDoc - si.GetDelCount()

	reader.fieldInfos, err = reader.initFieldInfos()
	if err != nil {
		return nil, err
	}
	reader.docValuesProducer, err = reader.initDocValuesProducer()
	if err != nil {
		return nil, err
	}
	return reader, nil
}

func NewSegmentReaderV1(si *SegmentCommitInfo, sr *SegmentReader,
	liveDocs, hardLiveDocs util.Bits, numDocs int, isNRT bool) (*SegmentReader, error) {

	panic("")
}

func (s *SegmentReader) Directory() store.Directory {
	return s.si.info.dir
}

func (s *SegmentReader) NumDocs() int {
	// Don't call ensureOpen() here (it could affect performance)
	return s.numDocs
}

func (s *SegmentReader) MaxDoc() int {
	// Don't call ensureOpen() here (it could affect performance)
	v, _ := s.si.info.MaxDoc()
	return v
}

func (s *SegmentReader) DoClose() error {
	if err := s.core.DecRef(); err != nil {
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

func (s *SegmentReader) GetReaderCacheHelper() CacheHelper {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentReader) GetFieldInfos() *FieldInfos {
	return s.fieldInfos
}

func (s *SegmentReader) GetLiveDocs() util.Bits {
	return s.liveDocs
}

func (s *SegmentReader) GetMetaData() *LeafMetaData {
	return s.metaData
}

func (s *SegmentReader) GetFieldsReader() StoredFieldsReader {
	return s.core.fieldsReaderLocal
}

func (s *SegmentReader) GetTermVectorsReader() TermVectorsReader {
	//ensureOpen();
	return s.core.termVectorsLocal
}

func (s *SegmentReader) GetNormsReader() NormsProducer {
	return s.core.normsProducer
}

func (s *SegmentReader) GetDocValuesReader() DocValuesProducer {
	//ensureOpen();
	return s.docValuesProducer
}

func (s *SegmentReader) GetPostingsReader() FieldsProducer {
	//ensureOpen();
	return s.core.fields
}

func (s *SegmentReader) GetPointsReader() PointsReader {
	return s.core.pointsReader
}

// GetOriginalSegmentInfo
// Returns the original SegmentInfo passed to the segment reader on creation time.
// getSegmentInfo() returns a clone of this instance.
func (s *SegmentReader) GetOriginalSegmentInfo() *SegmentCommitInfo {
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

func (s *SegmentReader) initFieldInfos() (*FieldInfos, error) {
	if !s.si.HasFieldUpdates() {
		return s.core.coreFieldInfos, nil
	} else {
		// updates always outside of CFS
		fisFormat := s.si.info.GetCodec().FieldInfosFormat()
		segmentSuffix := strconv.FormatInt(s.si.GetFieldInfosGen(), 36)
		return fisFormat.Read(s.si.info.dir, s.si.info, segmentSuffix, nil)
	}
}

// init most recent DocValues for the current commit
func (s *SegmentReader) initDocValuesProducer() (DocValuesProducer, error) {
	if s.fieldInfos.HasDocValues() == false {
		return nil, nil
	} else {
		var dir store.Directory
		if s.core.cfsReader != nil {
			dir = s.core.cfsReader
		} else {
			dir = s.si.info.dir
		}
		if s.si.HasFieldUpdates() {
			return NewSegmentDocValuesProducer(s.si, dir, s.core.coreFieldInfos, s.fieldInfos, s.segDocValues)
		} else {
			// simple case, no DocValues updates
			return s.segDocValues.GetDocValuesProducer(-1, s.si, dir, s.fieldInfos)
		}
	}
}
