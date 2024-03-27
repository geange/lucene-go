package index

import (
	"context"
	"errors"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/version"
	"github.com/matishsiao/goInfo"
	"golang.org/x/exp/maps"
	"runtime"
	"strconv"
	"time"
)

var (
	osInfo goInfo.GoInfoObject
)

func init() {
	info, err := goInfo.GetInfo()
	if err != nil {
		panic(err)
	}
	osInfo = info
}

func SetDiagnostics(info *SegmentInfo, source string, details map[string]string) error {
	diagnostics := make(map[string]string)

	diagnostics["source"] = source
	diagnostics["lucene.version"] = version.Last.String()
	diagnostics["os"] = runtime.GOOS
	diagnostics["os.arch"] = runtime.GOARCH
	diagnostics["os.version"] = ""
	diagnostics["go.version"] = runtime.Version()
	diagnostics["timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	maps.Copy(diagnostics, details)
	info.SetDiagnostics(diagnostics)
	return nil
}

func CreateCompoundFile(ctx context.Context, directory *store.TrackingDirectoryWrapper, info *SegmentInfo,
	ioContext *store.IOContext, deleteFiles func(files map[string]struct{})) error {

	// maybe this check is not needed, but why take the risk?
	if len(directory.GetCreatedFiles()) != 0 {
		return errors.New("pass a clean trackingdir for CFS creation")
	}

	if err := info.GetCodec().CompoundFormat().Write(ctx, directory, info, ioContext); err != nil {
		// Safe: these files must exist
		deleteFiles(directory.GetCreatedFiles())
		return err
	}

	// Replace all previous files with the CFS/CFE files:
	info.SetFiles(directory.GetCreatedFiles())

	return nil
}

var _ FlushNotifications = &flushNotifications{}

func (w *IndexWriter) newFlushNotifications() FlushNotifications {
	return &flushNotifications{
		w:          w,
		eventQueue: w.eventQueue,
	}
}

type flushNotifications struct {
	w          *IndexWriter
	eventQueue *EventQueue
}

func (f *flushNotifications) DeleteUnusedFiles(files map[string]struct{}) {
	f.eventQueue.Add(func(w *IndexWriter) error {
		return w.deleteNewFiles(files)
	})
}

func (f *flushNotifications) FlushFailed(info *SegmentInfo) {
	f.eventQueue.Add(func(w *IndexWriter) error {
		files := info.Files()
		return w.deleteNewFiles(files)
	})
}

func (f *flushNotifications) AfterSegmentsFlushed() error {
	return f.w.publishFlushedSegments(false)
}

func (f *flushNotifications) OnDeletesApplied() {
	f.eventQueue.Add(func(w *IndexWriter) error {
		_ = w.publishFlushedSegments(true)
		f.w.flushCount.Add(1)
		return nil
	})
}

func (f *flushNotifications) OnTicketBacklog() {
	f.eventQueue.Add(func(w *IndexWriter) error {
		return w.publishFlushedSegments(true)
	})
}
