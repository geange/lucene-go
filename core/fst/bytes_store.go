package fst

import (
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
)

var _ store.DataOutput = &BytesStore{}

// BytesStore
// TODO: merge with PagedBytes, except PagedBytes doesn't
// let you read while writing which FST needs
type BytesStore struct {
	*store.DataOutputImp

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

	this.DataOutputImp = store.NewDataOutputImp(this)

	return this
}

func NewBytesStore3(in store.DataInput, numBytes, maxBlockSize int) (*BytesStore, error) {
	blockSize := 2
	blockBits := 1
	for blockSize < numBytes && blockSize < maxBlockSize {
		blockSize *= 2
		blockBits++
	}

	this := &BytesStore{
		blockSize: blockSize,
		blockBits: blockBits,
		blockMask: blockSize - 1,
	}

	left := numBytes
	for left > 0 {
		chunk := util.Min(blockSize, left)
		block := make([]byte, chunk)
		err := in.ReadBytes(block)
		if err != nil {
			return nil, err
		}

		this.blocks = append(this.blocks, block)
		left -= chunk
	}

	// So .getPosition still works
	this.nextWrite = len(this.blocks[len(this.blocks)-1])
	return this, nil
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

func (r *BytesStore) getBlockBits() int {
	return r.blockBits
}

func (r *BytesStore) writeByte(dest int, b byte) {
	blockIndex := (int)(dest >> r.blockBits)
	block := r.blocks[blockIndex]
	block[dest&r.blockMask] = b
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
func (r *BytesStore) SkipBytes(size int) error {
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
	return nil
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

var _ BytesReader = &forwardReader{}

type forwardReader struct {
	*store.DataInputImp

	parent *BytesStore

	current    []byte
	nextBuffer int
	nextRead   int
}

func (r *forwardReader) ReadByte() (byte, error) {
	if r.nextRead == r.parent.blockSize {
		r.current = r.parent.blocks[r.nextBuffer]
		r.nextBuffer++
		r.nextRead = 0
	}
	v := r.current[r.nextRead]
	r.nextRead++
	return v, nil
}

func (r *forwardReader) ReadBytes(b []byte) error {
	offset, size := 0, len(b)

	for size > 0 {
		chunkLeft := r.parent.blockSize - r.nextRead
		if size <= chunkLeft {
			copy(b[offset:], r.current[r.nextRead:r.nextRead+size])
			r.nextRead += size
			break
		} else {
			if chunkLeft > 0 {
				copy(b[offset:], r.current[r.nextRead:r.nextBuffer+chunkLeft])
				offset += chunkLeft
				size -= chunkLeft
			}
			r.current = r.parent.blocks[r.nextBuffer]
			r.nextBuffer++
			r.nextRead = 0
		}
	}
	return nil
}

func (r *forwardReader) GetPosition() int64 {
	return int64((r.nextBuffer-1)*r.parent.blockSize + r.nextRead)
}

func (r *forwardReader) SetPosition(pos int64) error {
	bufferIndex := (int)(pos >> r.parent.blockBits)
	if r.nextBuffer != bufferIndex+1 {
		r.nextBuffer = bufferIndex + 1
		r.current = r.parent.blocks[bufferIndex]
	}
	r.nextRead = int(pos) & r.parent.blockMask
	return nil
}

func (r *forwardReader) Reversed() bool {
	return false
}

func (r *BytesStore) GetForwardReader() BytesReader {
	reader := &forwardReader{
		current:    nil,
		nextBuffer: 0,
		nextRead:   r.blockSize,
		parent:     r,
	}
	reader.DataInputImp = store.NewDataInputImp(reader)
	return reader
}

var _ BytesReader = &reverseReader{}

type reverseReader struct {
	*store.DataInputImp
	parent *BytesStore

	current    []byte
	nextBuffer int
	nextRead   int
}

func (r *reverseReader) GetPosition() int64 {
	return int64((r.nextBuffer+1)*r.parent.blockSize + r.nextRead)
}

func (r *reverseReader) ReadByte() (byte, error) {
	if r.nextRead == -1 {
		r.current = r.parent.blocks[r.nextBuffer]
		r.nextBuffer--
		r.nextRead = r.parent.blockSize - 1
	}
	v := r.current[r.nextRead]
	r.nextRead--
	return v, nil
}

func (r *reverseReader) ReadBytes(b []byte) error {
	for i := 0; i < len(b); i++ {
		v, err := r.ReadByte()
		if err != nil {
			return err
		}
		b[i] = v
	}
	return nil
}

func (r *reverseReader) SetPosition(pos int64) error {
	// NOTE: a little weird because if you
	// setPosition(0), the next byte you read is
	// bytes[0] ... but I would expect bytes[-1] (ie,
	// EOF)...?
	bufferIndex := (int)(pos >> r.parent.blockBits)
	if r.nextBuffer != bufferIndex-1 {
		r.nextBuffer = bufferIndex - 1
		r.current = r.parent.blocks[bufferIndex]
	}
	r.nextRead = int(pos) & r.parent.blockMask
	return nil
}

func (r *reverseReader) Reversed() bool {
	return true
}

func (r *BytesStore) GetReverseReader() BytesReader {
	return r.getReverseReader(true)
}

func (r *BytesStore) getReverseReader(allowSingle bool) BytesReader {
	if allowSingle && len(r.blocks) == 1 {
		return NewReverseBytesReader(r.blocks[0])
	}
	return r.newReverseReader()
}

func (r *BytesStore) newReverseReader() BytesReader {
	var current []byte
	if len(r.blocks) > 0 {
		current = r.blocks[0]
	}

	reader := &reverseReader{
		parent:     r,
		current:    current,
		nextBuffer: -1,
		nextRead:   0,
	}
	reader.DataInputImp = store.NewDataInputImp(reader)
	return reader
}
