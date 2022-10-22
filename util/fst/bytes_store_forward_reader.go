package fst

import "github.com/geange/lucene-go/core/store"

var _ BytesReader = &bytesStoreForwardReader{}

type bytesStoreForwardReader struct {
	*store.DataInputImp

	current    []byte
	nextBuffer int
	nextRead   int
	ptr        *BytesStore
}

func (bs *BytesStore) newBytesStoreForwardReader() *bytesStoreForwardReader {
	input := &bytesStoreForwardReader{
		nextRead: bs.blockSize,
		ptr:      bs,
	}
	input.DataInputImp = store.NewDataInputImp(input)
	return input
}

func (r *bytesStoreForwardReader) ReadByte() (byte, error) {
	if r.nextRead == r.ptr.blockSize {
		block, err := r.ptr.blocks.Get(r.nextBuffer)
		if err != nil {
			return 0, err
		}
		r.nextBuffer++
		r.current = block
		r.nextRead = 0
	}
	b := r.current[r.nextRead]
	r.nextRead++
	return b, nil
}

func (r *bytesStoreForwardReader) ReadBytes(b []byte) error {
	offset, size := 0, len(b)
	for size > 0 {
		chunk := r.ptr.blockSize - r.nextRead
		if size <= chunk {
			copy(b, r.current[r.nextRead:r.nextRead+size])
			r.nextRead += size
			break
		} else {
			if chunk > 0 {
				copy(b[offset:], r.current[r.nextRead:chunk+r.nextRead])
				offset += chunk
				size -= chunk
			}
			block, err := r.ptr.blocks.Get(r.nextBuffer)
			if err != nil {
				return err
			}
			r.current = block
			r.nextRead = 0
		}
	}
	return nil
}

func (r *bytesStoreForwardReader) SkipBytes(count int) error {
	r.SetPosition(r.GetPosition() + count)
	return nil
}

func (r *bytesStoreForwardReader) GetPosition() int {
	return int(int64((r.nextBuffer-1)*r.ptr.blockSize + r.nextRead))
}

func (r *bytesStoreForwardReader) SetPosition(pos int) {
	bufferIndex := pos >> r.ptr.blockBits
	if r.nextBuffer != bufferIndex+1 {
		r.nextBuffer = bufferIndex + 1
		block, err := r.ptr.blocks.Get(bufferIndex)
		if err != nil {
			return
		}
		r.current = block
	}
	r.nextRead = pos & r.ptr.blockMask
}

func (r *bytesStoreForwardReader) Reversed() bool {
	return false
}
