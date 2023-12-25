package bkd

import (
	"encoding/binary"
	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/store"
	"io"
)

var _ PointReader = &OfflinePointReader{}

type OfflinePointReader struct {
	countLeft      int
	in             store.IndexInput
	onHeapBuffer   []byte
	offset         int
	checked        bool
	config         *Config
	pointsInBuffer int
	maxPointOnHeap int
	name           string // File name we are reading
	pointValue     *OfflinePointValue
}

func NewOfflinePointReader(config *Config, tempDir store.Directory,
	tempFileName string, start, length int, reusableBuffer []byte) (*OfflinePointReader, error) {

	reader := &OfflinePointReader{
		countLeft:      0,
		in:             nil,
		onHeapBuffer:   nil,
		offset:         0,
		checked:        false,
		config:         config,
		pointsInBuffer: 0,
		maxPointOnHeap: len(reusableBuffer) / config.BytesPerDoc(),
		name:           tempFileName,
		pointValue:     nil,
	}

	// TODO: need impl it ?
	//     if ((start + length) * config.BytesPerDoc() + CodecUtil.footerLength() > tempDir.fileLength(tempFileName)) {
	//      throw new IllegalArgumentException("requested slice is beyond the length of this file: start=" + start + " length=" + length + " BytesPerDoc=" + config.BytesPerDoc() + " fileLength=" + tempDir.fileLength(tempFileName) + " tempFileName=" + tempFileName);
	//    }
	//    if (reusableBuffer == null) {
	//      throw new IllegalArgumentException("[reusableBuffer] cannot be null");
	//    }
	//    if (reusableBuffer.length < config.BytesPerDoc()) {
	//      throw new IllegalArgumentException("Len of [reusableBuffer] must be bigger than " + config.BytesPerDoc());
	//    }
	// Best-effort checksumming:
	fileLength, err := tempDir.FileLength(nil, tempFileName)
	if err != nil {
		return nil, err
	}

	if start == 0 && length*config.BytesPerDoc() == int(fileLength)-utils.FooterLength() {

		// If we are going to read the entire file, e.g. because BKDWriter is now
		// partitioning it, we open with checksums:
		reader.in, err = store.OpenChecksumInput(tempDir, tempFileName)
		if err != nil {
			return nil, err
		}
	} else {
		// Since we are going to seek somewhere in the middle of a possibly huge
		// file, and not read all bytes from there, don't use ChecksumIndexInput here.
		// This is typically fine, because this same file will later be read fully,
		// at another level of the BKDWriter recursion
		reader.in, err = tempDir.OpenInput(nil, tempFileName)
		if err != nil {
			return nil, err
		}
	}

	seekFP := start * (config.BytesPerDoc())
	if _, err = reader.in.Seek(int64(seekFP), io.SeekStart); err != nil {
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
	if r.pointsInBuffer > 0 {
		r.pointsInBuffer--
		r.offset += r.config.BytesPerDoc()
		return true, nil
	}

	if r.countLeft == 0 {
		return false, nil
	}

	if r.countLeft > r.maxPointOnHeap {
		size := r.maxPointOnHeap * r.config.BytesPerDoc()
		_, err := r.in.Read(r.onHeapBuffer[0:size])
		if err != nil {
			return false, err
		}
		r.pointsInBuffer = r.maxPointOnHeap - 1
		r.countLeft -= r.maxPointOnHeap
	} else {
		size := r.countLeft * r.config.BytesPerDoc()
		_, err := r.in.Read(r.onHeapBuffer[0:size])
		if err != nil {
			return false, err
		}
		r.pointsInBuffer = r.countLeft - 1
		r.countLeft = 0
	}
	r.offset = 0
	return true, nil
}

func (r *OfflinePointReader) PointValue() PointValue {
	r.pointValue.SetOffset(r.offset)
	return r.pointValue
}

var _ PointValue = &OfflinePointValue{}

// OfflinePointValue Reusable implementation for a point value offline
type OfflinePointValue struct {
	config *Config
	bytes  []byte
	offset int
}

func NewOfflinePointValue(config *Config, value []byte) *OfflinePointValue {
	return &OfflinePointValue{
		config: config,
		bytes:  value,
	}
}

func (v *OfflinePointValue) PackedValue() []byte {
	return v.bytes[v.offset : v.offset+v.config.PackedBytesLength()]
}

func (v *OfflinePointValue) DocID() int {
	position := v.offset + v.config.PackedBytesLength()
	return int(binary.BigEndian.Uint32(v.bytes[position:]))
}

func (v *OfflinePointValue) PackedValueDocIDBytes() []byte {
	return v.bytes[v.offset : v.offset+v.config.BytesPerDoc()]
}

// SetOffset Sets a new value by changing the offset.
func (v *OfflinePointValue) SetOffset(offset int) {
	v.offset = offset
}
