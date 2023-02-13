package bkd

import (
	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/store"
)

var _ PointWriter = &OfflinePointWriter{}

// OfflinePointWriter Writes points to disk in a fixed-with format.
// lucene.internal
type OfflinePointWriter struct {
	tempDir       store.Directory
	out           store.IndexOutput
	name          string
	config        *BKDConfig
	count         int64
	closed        bool
	expectedCount int64
}

func NewOfflinePointWriter(config *BKDConfig, tempDir store.Directory,
	tempFileNamePrefix, desc string, expectedCount int64) *OfflinePointWriter {
	out, err := tempDir.CreateTempOutput(tempFileNamePrefix, "bkd_"+desc, nil)
	if err != nil {
		return nil
	}

	return &OfflinePointWriter{
		tempDir:       tempDir,
		out:           out,
		name:          out.GetName(),
		config:        config,
		count:         0,
		closed:        false,
		expectedCount: expectedCount,
	}
}

func (w *OfflinePointWriter) Close() error {
	if err := utils.WriteFooter(w.out); err != nil {
		return err
	}

	if err := w.out.Close(); err != nil {
		return err
	}
	w.closed = true
	return nil
}

func (w *OfflinePointWriter) Append(packedValue []byte, docID int) error {
	// TODO: need impl it ?
	//assert closed == false : "Point writer is already closed";
	//assert packedValue.length == config.packedBytesLength : "[packedValue] must have length [" + config.packedBytesLength + "] but was [" + packedValue.length + "]";
	if _, err := w.out.Write(packedValue); err != nil {
		return err
	}
	if err := w.out.WriteUint32(uint32(docID)); err != nil {
		return err
	}
	w.count++
	return nil
	//assert expectedCount == 0 || count <= expectedCount:  "expectedCount=" + expectedCount + " vs count=" + count;
}

func (w *OfflinePointWriter) AppendValue(pointValue PointValue) error {
	//assert closed == false : "Point writer is already closed";
	packedValueDocID := pointValue.PackedValueDocIDBytes()
	//assert packedValueDocID.length == config.bytesPerDoc : "[packedValue and docID] must have length [" + (config.bytesPerDoc) + "] but was [" + packedValueDocID.length + "]";
	if _, err := w.out.Write(packedValueDocID); err != nil {
		return err
	}
	w.count++
	return nil
	//assert expectedCount == 0 || count <= expectedCount : "expectedCount=" + expectedCount + " vs count=" + count;
}

func (w *OfflinePointWriter) GetReader(start, length int64) (PointReader, error) {
	buffer := make([]byte, w.config.BytesPerDoc)
	return w.getReader(start, length, buffer)
}

func (w *OfflinePointWriter) getReader(start, length int64, reusableBuffer []byte) (*OfflinePointReader, error) {
	// TODO: need impl it ?
	//assert closed: "point writer is still open and trying to get a reader";
	//assert start + length <= count: "start=" + start + " length=" + length + " count=" + count;
	//assert expectedCount == 0 || count == expectedCount;
	return NewOfflinePointReader(w.config, w.tempDir, w.name, start, length, reusableBuffer)
}

func (w *OfflinePointWriter) Count() int64 {
	return w.count
}

func (w *OfflinePointWriter) Destroy() error {
	return w.tempDir.DeleteFile(w.name)
}

func (w *OfflinePointWriter) GetIndexOutput() store.IndexOutput {
	return w.out
}
