package bkd

import (
	"encoding/binary"
	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	"io"
)

var _ PointReader = &OfflinePointReader{}

type OfflinePointReader struct {
	countLeft      int64
	in             store.IndexInput
	onHeapBuffer   []byte
	offset         int
	checked        bool
	config         *BKDConfig
	pointsInBuffer int
	maxPointOnHeap int
	name           string // File name we are reading
	pointValue     *OfflinePointValue
}

func NewOfflinePointReader(config *BKDConfig, tempDir store.Directory,
	tempFileName string, start, length int64, reusableBuffer []byte) (*OfflinePointReader, error) {

	reader := &OfflinePointReader{
		countLeft:      0,
		in:             nil,
		onHeapBuffer:   nil,
		offset:         0,
		checked:        false,
		config:         config,
		pointsInBuffer: 0,
		maxPointOnHeap: len(reusableBuffer) / config.BytesPerDoc,
		name:           "",
		pointValue:     nil,
	}

	// TODO: need impl it ?
	//     if ((start + length) * config.bytesPerDoc + CodecUtil.footerLength() > tempDir.fileLength(tempFileName)) {
	//      throw new IllegalArgumentException("requested slice is beyond the length of this file: start=" + start + " length=" + length + " bytesPerDoc=" + config.bytesPerDoc + " fileLength=" + tempDir.fileLength(tempFileName) + " tempFileName=" + tempFileName);
	//    }
	//    if (reusableBuffer == null) {
	//      throw new IllegalArgumentException("[reusableBuffer] cannot be null");
	//    }
	//    if (reusableBuffer.length < config.bytesPerDoc) {
	//      throw new IllegalArgumentException("Length of [reusableBuffer] must be bigger than " + config.bytesPerDoc);
	//    }
	// Best-effort checksumming:
	fileLength, err := tempDir.FileLength(tempFileName)
	if err != nil {
		return nil, err
	}

	if start == 0 && length*int64(config.BytesPerDoc) == (fileLength)-int64(utils.FooterLength()) {

		// If we are going to read the entire file, e.g. because BKDWriter is now
		// partitioning it, we open with checksums:
		reader.in, err = store.OpenChecksumInput(tempDir, tempFileName, nil)
		if err != nil {
			return nil, err
		}
	} else {
		// Since we are going to seek somewhere in the middle of a possibly huge
		// file, and not read all bytes from there, don't use ChecksumIndexInput here.
		// This is typically fine, because this same file will later be read fully,
		// at another level of the BKDWriter recursion
		reader.in, err = tempDir.OpenInput(tempFileName, nil)
		if err != nil {
			return nil, err
		}
	}

	reader.name = tempFileName
	seekFP := start * int64(config.BytesPerDoc)
	if _, err = reader.in.Seek(seekFP, io.SeekStart); err != nil {
		return nil, err
	}
	reader.countLeft = length
	reader.onHeapBuffer = reusableBuffer
	reader.pointValue = NewOfflinePointValue(config, reader.onHeapBuffer)
	return reader, nil
}

func (r *OfflinePointReader) Close() error {
	if r.countLeft == 0 && r.checked == false {
		if in, ok := r.in.(store.ChecksumIndexInput); ok {
			r.checked = true
			if err := utils.CheckFooter(in); err != nil {
				return err
			}
		}
	}
	return r.in.Close()
}

func (r *OfflinePointReader) Next() (bool, error) {
	if r.pointsInBuffer == 0 {
		if r.countLeft == 0 {
			return false, nil
		}

		if int(r.countLeft) > r.maxPointOnHeap {
			size := r.maxPointOnHeap * r.config.BytesPerDoc
			_, err := r.in.Read(r.onHeapBuffer[0:size])
			if err != nil {
				return false, err
			}
			r.pointsInBuffer = r.maxPointOnHeap - 1
			r.countLeft -= int64(r.maxPointOnHeap)
		} else {
			size := int(r.countLeft) * r.config.BytesPerDoc
			_, err := r.in.Read(r.onHeapBuffer[0:size])
			if err != nil {
				return false, err
			}
			r.pointsInBuffer = int(r.countLeft - 1)
			r.countLeft = 0
		}
		r.offset = 0
	} else {
		r.pointsInBuffer--
		r.offset += r.config.BytesPerDoc
	}
	return true, nil
}

func (r *OfflinePointReader) PointValue() PointValue {
	r.pointValue.SetOffset(r.offset)
	return r.pointValue
}

var _ PointValue = &OfflinePointValue{}

// OfflinePointValue Reusable implementation for a point value offline
type OfflinePointValue struct {
	packedValue       *util.BytesRef
	packedValueDocID  *util.BytesRef
	packedValueLength int
}

func NewOfflinePointValue(config *BKDConfig, value []byte) *OfflinePointValue {
	return &OfflinePointValue{
		packedValue:       util.NewBytesRef(value, 0, config.PackedBytesLength),
		packedValueDocID:  util.NewBytesRef(value, 0, config.BytesPerDoc),
		packedValueLength: config.PackedBytesLength,
	}
}

func (v *OfflinePointValue) PackedValue() []byte {
	return v.packedValue.GetBytes()
}

func (v *OfflinePointValue) DocID() int {
	position := v.packedValueDocID.Offset + v.packedValueLength
	return int(binary.BigEndian.Uint32(v.packedValueDocID.GetBytes()[position:]))
}

func (v *OfflinePointValue) PackedValueDocIDBytes() []byte {
	return v.packedValueDocID.GetBytes()
}

// SetOffset Sets a new value by changing the offset.
func (v *OfflinePointValue) SetOffset(offset int) {
	v.packedValue.Offset = offset
	v.packedValueDocID.Offset = offset
}
