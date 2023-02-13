package index

import (
	"github.com/geange/lucene-go/core/store"
	"go.uber.org/atomic"
	"io"
	"sync"
)

const (
	INFO_VERBOSE = false
)

type DocumentsWriterPerThread struct {
	codec     Codec
	directory *store.TrackingDirectoryWrapper
	consumer  DocConsumer

	// Updates for our still-in-RAM (to be flushed next) segment
	pendingUpdates *BufferedUpdates
	// Current segment we are working on
	segmentInfo            *SegmentInfo
	aborted                bool
	flushPending           bool
	lastCommittedBytesUsed int64
	hasFlushed             bool
	fieldInfos             *FieldInfosBuilder
	infoStream             io.Writer
	numDocsInRAM           int
	deleteQueue            *DocumentsWriterDeleteQueue
	deleteSlice            *DeleteSlice
	pendingNumDocs         *atomic.Int64
	indexWriterConfig      *LiveIndexWriterConfig
	enableTestPoints       bool
	lock                   sync.RWMutex
	deleteDocIDs           []int
	numDeletedDocIds       int
}

type IndexingChain interface {
	GetChain(indexCreatedVersionMajor int, segmentInfo *SegmentInfo, directory store.Directory,
		fieldInfos *FieldInfosBuilder, indexWriterConfig *LiveIndexWriterConfig)
}
