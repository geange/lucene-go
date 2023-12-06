package bkd

import (
	"context"
	"errors"
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
	config        *Config
	count         int
	closed        bool
	expectedCount int
}

func NewOfflinePointWriter(config *Config, tempDir store.Directory,
	tempFileNamePrefix, desc string, expectedCount int) *OfflinePointWriter {
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
	if w.closed {
		return nil
	}

	if err := utils.WriteFooter(w.out); err != nil {
		return err
	}

	if err := w.out.Close(); err != nil {
		return err
	}
	w.closed = true
	return nil
}

func (w *OfflinePointWriter) Append(ctx context.Context, packedValue []byte, docID int) error {
	if _, err := w.out.Write(packedValue); err != nil {
		return err
	}
	if err := w.out.WriteUint32(ctx, uint32(docID)); err != nil {
		return err
	}
	w.count++
	return nil
}

func (w *OfflinePointWriter) AppendPoint(pointValue PointValue) error {
	//assert closed == false : "Point writer is already closed";
	packedValueDocID := pointValue.PackedValueDocIDBytes()
	//assert packedValueDocID.length == config.BytesPerDoc : "[packedValue and docID] must have length [" + (config.BytesPerDoc) + "] but was [" + packedValueDocID.length + "]";
	if _, err := w.out.Write(packedValueDocID); err != nil {
		return err
	}
	w.count++
	return nil
	//assert expectedCount == 0 || count <= expectedCount : "expectedCount=" + expectedCount + " vs count=" + count;
}

func (w *OfflinePointWriter) GetReader(startPoint, length int) (PointReader, error) {
	if !w.closed {
		return nil, errors.New("point writer is still open and trying to get a reader")
	}

	buffer := make([]byte, w.config.BytesPerDoc())
	return w.getReader(startPoint, length, buffer)
}

func (w *OfflinePointWriter) getReader(start, length int, reusableBuffer []byte) (*OfflinePointReader, error) {
	// TODO: need impl it ?
	//assert closed: "point writer is still open and trying to get a reader";
	//assert start + length <= count: "start=" + start + " length=" + length + " count=" + count;
	//assert expectedCount == 0 || count == expectedCount;
	return NewOfflinePointReader(w.config, w.tempDir, w.name, start, length, reusableBuffer)
}

func (w *OfflinePointWriter) Count() int {
	return w.count
}

func (w *OfflinePointWriter) Destroy() error {
	return w.tempDir.DeleteFile(w.name)
}

func (w *OfflinePointWriter) GetIndexOutput() store.IndexOutput {
	return w.out
}
