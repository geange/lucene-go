package fst

import (
	"github.com/geange/lucene-go/core/store"
)

var _ store.DataOutput = &BytesStore{}

// BytesStore
// TODO: merge with PagedBytes, except PagedBytes doesn't
// let you read while writing which FST needs
type BytesStore struct {
	blocks    [][]byte
	blockSize int
	blockBits int
	blockMask int
	current   []byte
	nextWrite int
}

func NewBytesStore(blockBits int) *BytesStore {
	this := &BytesStore{
		blockBits: blockBits,
	}
	this.blockSize = 1 << blockBits
	this.blockMask = this.blockSize - 1
	this.nextWrite = this.blockSize
	return this
}

func (r *BytesStore) WriteByteIndex(dest int64, v byte) {
	blockIndex := (int)(dest >> r.blockBits)
	block := r.blocks[blockIndex]
	block[int(dest)&r.blockMask] = v
}

func (r *BytesStore) WriteByte(v byte) error {
	if r.nextWrite == r.blockSize {
		r.current = make([]byte, r.blockSize)
		r.blocks = append(r.blocks, r.current)
		r.nextWrite = 0
	}
	r.current[r.nextWrite] = v
	r.nextWrite++
	return nil
}

func (r *BytesStore) WriteBytes(bs []byte) error {
	offset := 0
	size := len(bs)
	for size > 0 {
		chunk := r.blockSize - r.nextWrite
		if size <= chunk {
			copy(r.current[r.nextWrite:r.nextWrite+size], bs)
			r.nextWrite += size
			break
		} else {
			if chunk > 0 {
				copy(r.current[r.nextWrite:], bs[offset:])
				offset += chunk
				size -= chunk
			}
			r.current = make([]byte, r.blockSize)
			r.blocks = append(r.blocks, r.current)
			r.nextWrite = 0
		}
	}
	return nil
}

func (r *BytesStore) CopyBytes(input store.DataInput, numBytes int) error {
	return nil
}

func (r *BytesStore) getBlockBits() int {
	return r.blockBits
}

// Absolute writeBytes without changing the current position. Note: this cannot "grow" the bytes,
// so you must only call it on already written parts.
// 该复制从尾部向前复制
func (r *BytesStore) writeBytes(dest int, bs []byte) {
	// Note: weird: must go "backwards" because copyBytes
	// calls us with overlapping src/dest.  If we
	// go forwards then we overwrite bytes before we can
	// copy them:

	size := len(bs)
	offset := 0

	end := dest + size
	blockIndex := end >> r.blockBits
	downTo := end & r.blockMask
	if downTo == 0 {
		blockIndex--
		downTo = r.blockSize
	}
	block := r.blocks[blockIndex]

	for size > 0 {
		if size <= downTo {
			copy(block[downTo-size:downTo], bs)
			break
		} else {
			size -= downTo
			copy(block[0:downTo], bs[offset+size:offset+size+downTo])
			blockIndex--
			block = r.blocks[blockIndex]
			downTo = r.blockSize
		}
	}
}

// CopyBytesToSelf Absolute copy bytes self to self, without changing the position. Note: this cannot "grow" the bytes,
// so must only call it on already written parts.
func (r *BytesStore) CopyBytesToSelf(src, dest, size int) {
	// Note: weird: must go "backwards" because copyBytes
	// calls us with overlapping src/dest.  If we
	// go forwards then we overwrite bytes before we can
	// copy them:

	end := src + size
	blockIndex := end >> r.blockBits
	downTo := end & r.blockMask
	if downTo == 0 {
		blockIndex--
		downTo = r.blockSize
	}
	block := r.blocks[blockIndex]

	for size > 0 {
		if size <= downTo {
			r.writeBytes(dest, block[downTo-size:downTo])
			break
		} else {
			size -= downTo
			r.writeBytes(dest+size, block[0:downTo])
			blockIndex--
			block = r.blocks[blockIndex]
			downTo = r.blockSize
		}
	}
}

// CopyBytesToArray Copies bytes from this store to a target byte array.
func (r *BytesStore) CopyBytesToArray(src int, dest []byte) {
	offset, size := 0, len(dest)

	blockIndex := src >> r.blockBits
	upto := src & r.blockMask
	block := r.blocks[blockIndex]

	for size > 0 {
		chunk := r.blockSize - upto
		if size <= chunk {
			copy(dest, block[upto:])
			break
		} else {
			copy(dest, block[upto:])
			blockIndex++
			block = r.blocks[blockIndex]
			upto = 0
			size -= chunk
			offset += chunk
		}
	}
}

// WriteInt Writes an int at the absolute position without changing the current pointer.
func (r *BytesStore) WriteInt(pos int, value int) {
	blockIndex := pos >> r.blockBits
	upto := pos & r.blockMask
	block := r.blocks[blockIndex]
	shift := 24
	for i := 0; i < 4; i++ {
		block[upto] = byte(value >> shift)
		upto++
		shift -= 8
		if upto == r.blockSize {
			upto = 0
			blockIndex++
			block = r.blocks[blockIndex]
		}
	}
}

// Reverse from srcPos, inclusive, to destPos, inclusive.
// 将从 [srcPos, destPos] 的数据进行反转
func (r *BytesStore) Reverse(srcPos, destPos int) {
	srcBlockIndex := srcPos >> r.blockBits
	src := srcPos & r.blockMask
	srcBlock := r.blocks[srcBlockIndex]

	destBlockIndex := destPos >> r.blockBits
	dest := destPos & r.blockMask
	destBlock := r.blocks[destBlockIndex]

	limit := (destPos - srcPos + 1) / 2

	for i := 0; i < limit; i++ {
		b := srcBlock[src]
		srcBlock[src] = destBlock[dest]
		destBlock[dest] = b

		src++
		if src == r.blockSize {
			srcBlockIndex++
			srcBlock = r.blocks[srcBlockIndex]
			src = 0
		}

		dest--
		if dest == -1 {
			destBlockIndex--
			destBlock = r.blocks[destBlockIndex]
			dest = r.blockSize - 1
		}
	}
}

// SkipBytes 跳过 size 数量的字节
func (r *BytesStore) SkipBytes(size int) {
	for size > 0 {
		chunk := r.blockSize - r.nextWrite
		if size <= chunk {
			r.nextWrite += size
			break
		} else {
			size -= chunk
			r.current = make([]byte, r.blockSize)
			r.blocks = append(r.blocks, r.current)
			r.nextWrite = 0
		}
	}
}

func (r *BytesStore) GetPosition() int64 {
	return int64((len(r.blocks)-1)*r.blockSize + r.nextWrite)
}

// Truncate Pos must be less than the max position written so far! Ie, you cannot "grow" the file with this!
func (r *BytesStore) Truncate(newLen int) {
	blockIndex := (int)(newLen >> r.blockBits)
	r.nextWrite = newLen & r.blockMask
	if r.nextWrite == 0 {
		blockIndex--
		r.nextWrite = r.blockSize
	}

	r.blocks = r.blocks[:blockIndex+1]
	if newLen == 0 {
		r.current = nil
	} else {
		r.current = r.blocks[blockIndex]
	}
}

func (r *BytesStore) Finish() {
	if r.current != nil {
		lastBuffer := make([]byte, r.nextWrite)
		copy(lastBuffer, r.current[0:r.nextWrite])

		r.blocks = append(r.blocks, lastBuffer)
		r.current = nil
	}
}

// WriteTo Writes all of our bytes to the target DataOutput.
func (r *BytesStore) WriteTo(out store.DataOutput) error {
	for _, block := range r.blocks {
		if err := out.WriteBytes(block); err != nil {
			return err
		}
	}
	return nil
}
