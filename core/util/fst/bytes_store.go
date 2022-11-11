package fst

import (
	"fmt"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/math"
	"unsafe"
)

var (
	_ store.DataOutput = &BytesStore{}

	BytesStoreSize = unsafe.Sizeof(BytesStore{})
)

type BytesStore struct {
	blocks [][]byte

	blockSize int
	blockBits int
	blockMask int

	current   []byte
	nextWrite int

	*store.DataOutputImp
}

func NewBytesStore(blockBits int) *BytesStore {
	blockSize := 1 << blockBits

	bytesStore := &BytesStore{
		blocks:    make([][]byte, 0),
		blockSize: blockSize,
		blockBits: blockBits,
		blockMask: blockSize - 1,
		current:   make([]byte, 0),
		nextWrite: blockSize,
	}
	bytesStore.DataOutputImp = store.NewDataOutputImp(bytesStore)
	return bytesStore
}

// NewBytesStoreV1 Pulls bytes from the provided IndexInput.
func NewBytesStoreV1(in store.DataInput, numBytes int64, maxBlockSize int) (*BytesStore, error) {
	b := &BytesStore{}

	blockSize := 2
	blockBits := 1
	for blockSize < int(numBytes) && blockSize < maxBlockSize {
		blockSize *= 2
		blockBits++
	}
	b.blockBits = blockBits
	b.blockSize = blockSize
	b.blockMask = blockSize - 1
	left := int(numBytes)
	for left > 0 {
		chunk := math.Min(blockSize, left)
		block := make([]byte, chunk)
		err := in.ReadBytes(block)
		if err != nil {
			return nil, err
		}
		b.blocks = append(b.blocks, block)
		left -= chunk
	}

	// So .getPosition still works
	b.nextWrite = len(b.blocks[len(b.blocks)-1])
	return b, nil
}

// Absolute write byte; you must ensure dest is < max position written so far.
func (b *BytesStore) writeByte(dest int64, c byte) {
	blockIndex := (int)(dest >> b.blockBits)
	block := b.blocks[blockIndex]
	block[(int(dest) & b.blockMask)] = c
}

func (b *BytesStore) WriteBytes(bs []byte) error {
	offset, size := 0, len(bs)
	for size > 0 {
		chunk := b.blockSize - b.nextWrite
		if size <= chunk {
			//assert b != null;
			//assert current != null;
			copy(b.current[b.nextWrite:b.nextWrite+size], bs[offset:])
			b.nextWrite += size
			break
		} else {
			if chunk > 0 {
				copy(b.current[b.nextWrite:b.nextWrite+chunk], bs[offset:])
				offset += chunk
				size -= chunk
			}
			b.current = make([]byte, b.blockSize)
			b.blocks = append(b.blocks, b.current)
			b.nextWrite = 0
		}
	}
	return nil
}

func (b *BytesStore) getBlockBits() int {
	return b.blockBits
}

// Absolute writeBytes without changing the current position. Note: this cannot "grow" the bytes,
// so you must only call it on already written parts.
func (b *BytesStore) writeBytesAt(dest int64, bs []byte) error {
	offset, size := int64(0), int64(len(bs))

	err := assert(dest+size <= b.GetPosition())
	if err != nil {
		return err
	}

	end := dest + size
	blockIndex := end >> b.blockBits
	downTo := end & int64(b.blockMask)
	if downTo == 0 {
		blockIndex--
		downTo = int64(b.blockSize)
	}
	block := b.blocks[blockIndex]

	for size > 0 {
		//System.out.println("    cycle downTo=" + downTo + " len=" + len);
		if size <= downTo {
			copy(block[downTo-size:downTo], bs[offset:])
			break
		} else {
			size -= downTo
			copy(block[:downTo], bs[offset+size:])
			blockIndex--
			block = b.blocks[blockIndex]
			downTo = int64(b.blockSize)
		}
	}
	return nil
}

// MoveBytes bsolute copy bytes self to self, without changing the position. Note: this cannot "grow" the bytes,
// so must only call it on already written parts.
func (b *BytesStore) MoveBytes(src, dest int64, size int64) error {
	err := assert(src < dest)
	if err != nil {
		return err
	}

	// Note: weird: must go "backwards" because copyBytes
	// calls us with overlapping src/dest.  If we
	// go forwards then we overwrite bytes before we can
	// copy them:
	end := src + size

	blockIndex := end >> b.blockBits
	downTo := end & int64(b.blockMask)
	if downTo == 0 {
		blockIndex--
		downTo = int64(b.blockSize)
	}
	block := b.blocks[blockIndex]

	for size > 0 {
		if size <= downTo {
			err := b.writeBytesAt(dest, block[downTo-size:downTo])
			if err != nil {
				return err
			}
			break
		} else {
			size -= downTo
			err := b.writeBytesAt(dest+size, block[:downTo])
			if err != nil {
				return err
			}
			blockIndex--
			block = b.blocks[blockIndex]
			downTo = int64(b.blockSize)
		}
	}

	return nil
}

// CopyBytesToArray Copies bytes from this store to a target byte array.
func (b *BytesStore) CopyBytesToArray(src int64, dest []byte) error {
	blockIndex := src >> b.blockBits
	upto := src & int64(b.blockMask)
	block := b.blocks[blockIndex]

	offset := int64(0)
	size := int64(len(dest))

	for size > 0 {
		chunk := int64(b.blockSize) - upto
		if size <= chunk {
			copy(dest[offset:offset+size], block[upto:])
			break
		} else {
			copy(dest[offset:offset+chunk], block[upto:])
			blockIndex++
			block = b.blocks[blockIndex]
			upto = 0
			size -= chunk
			offset += chunk
		}
	}

	return nil
}

// WriteInt Writes an int at the absolute position without changing the current pointer.
func (b *BytesStore) WriteInt(pos int64, value int32) {
	blockIndex := pos >> b.blockBits
	upto := pos & int64(b.blockMask)
	block := b.blocks[blockIndex]
	shift := 24
	for i := 0; i < 4; i++ {
		block[upto] = byte(value >> shift)
		upto++
		shift -= 8
		if upto == int64(b.blockSize) {
			upto = 0
			blockIndex++
			block = b.blocks[blockIndex]
		}
	}
}

// Reverse from srcPos, inclusive, to destPos, inclusive.
func (b *BytesStore) Reverse(srcPos, destPos int64) error {
	err := assert(srcPos < destPos)
	if err != nil {
		return err
	}
	err = assert(destPos < b.GetPosition())
	if err != nil {
		return err
	}

	srcBlockIndex := srcPos >> b.blockBits
	src := srcPos & int64(b.blockMask)
	srcBlock := b.blocks[srcBlockIndex]

	destBlockIndex := destPos >> b.blockBits
	dest := destPos & int64(b.blockMask)
	destBlock := b.blocks[destBlockIndex]
	//System.out.println("  srcBlock=" + srcBlockIndex + " destBlock=" + destBlockIndex);

	limit := (destPos - srcPos + 1) / 2
	for i := int64(0); i < limit; i++ {
		//System.out.println("  cycle src=" + src + " dest=" + dest);
		v := srcBlock[src]
		srcBlock[src] = destBlock[dest]
		destBlock[dest] = v
		src++
		if src == int64(b.blockSize) {
			srcBlockIndex++
			srcBlock = b.blocks[srcBlockIndex]
			//System.out.println("  set destBlock=" + destBlock + " srcBlock=" + srcBlock);
			src = 0
		}

		dest--
		if dest == -1 {
			destBlockIndex--
			destBlock = b.blocks[destBlockIndex]
			//System.out.println("  set destBlock=" + destBlock + " srcBlock=" + srcBlock);
			dest = int64(b.blockSize) - 1
		}
	}

	return nil
}

func (b *BytesStore) SkipBytes(size int) {
	for size > 0 {
		chunk := b.blockSize - b.nextWrite
		if size <= chunk {
			b.nextWrite += size
			break
		} else {
			size -= chunk
			current := make([]byte, b.blockSize)
			b.blocks = append(b.blocks, current)
			b.nextWrite = 0
		}
	}
}

func (b *BytesStore) GetPosition() int64 {
	return int64((len(b.blocks)-1)*b.blockSize + b.nextWrite)
}

// Truncate Pos must be less than the max position written so far! Ie, you cannot "grow" the file with this!
func (b *BytesStore) Truncate(newLen int64) error {
	err := assert(newLen <= b.GetPosition())
	if err != nil {
		return err
	}
	err = assert(newLen >= 0)
	if err != nil {
		return err
	}

	blockIndex := newLen >> b.blockBits
	b.nextWrite = int(newLen & int64(b.blockMask))
	if b.nextWrite == 0 {
		blockIndex--
		b.nextWrite = b.blockSize
	}
	b.blocks = b.blocks[:blockIndex+1]
	if newLen == 0 {
		b.current = nil
	} else {
		b.current = b.blocks[blockIndex]
	}
	return assert(newLen == b.GetPosition())
}

func (b *BytesStore) Finish() {
	if b.current != nil {
		lastBuffer := make([]byte, b.nextWrite)
		copy(lastBuffer, b.current[:b.nextWrite])
		b.blocks[len(b.blocks)-1] = lastBuffer
		b.current = nil
	}
}

// WriteTo Writes all of our bytes to the target DataOutput.
func (b *BytesStore) WriteTo(out store.DataOutput) error {
	for _, block := range b.blocks {
		err := out.WriteBytes(block)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *BytesStore) GetReverseReader() BytesReader {
	return b.getReverseReader(true)
}

func (b *BytesStore) getReverseReader(allowSingle bool) BytesReader {
	if allowSingle && len(b.blocks) == 1 {
		return NewReverseBytesReader(b.blocks[0])
	}
	return newReverseReader(b)
}

func (b *BytesStore) RamBytesUsed() int64 {
	sum := int64(BytesStoreSize)
	for i := range b.blocks {
		sum += int64(cap(b.blocks[i]))
	}
	return sum
}

var _ BytesReader = &reverseReader{}

type reverseReader struct {
	store *BytesStore

	current    []byte
	nextBuffer int
	nextRead   int

	*store.DataInputImp
}

func newReverseReader(in *BytesStore) *reverseReader {
	reader := &reverseReader{
		store:      in,
		nextBuffer: -1,
		nextRead:   0,
	}

	if len(in.blocks) != 0 {
		reader.current = in.blocks[0]
	}

	reader.DataInputImp = store.NewDataInputImp(reader)
	return reader
}

func (r *reverseReader) ReadByte() (byte, error) {
	if r.nextRead == -1 {
		r.current = r.store.blocks[r.nextBuffer]
		r.nextBuffer--
		r.nextRead = r.store.blockSize - 1
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

func (r *reverseReader) GetPosition() int64 {
	return int64((r.nextBuffer+1)*r.store.blockSize + r.nextRead)
}

func (r *reverseReader) SetPosition(pos int64) {
	// NOTE: a little weird because if you
	// setPosition(0), the next byte you read is
	// bytes[0] ... but I would expect bytes[-1] (ie,
	// EOF)...?
	bufferIndex := int(pos >> r.store.blockBits)
	if r.nextBuffer != bufferIndex-1 {
		r.nextBuffer = bufferIndex - 1
		r.current = r.store.blocks[bufferIndex]
	}
	r.nextRead = int(pos & int64(r.store.blockMask))

	err := assert(r.GetPosition() == pos, fmt.Sprintf("pos=%d getPos()=%d", pos, r.GetPosition()))
	if err != nil {
		return
	}
}

func (r *reverseReader) Reversed() bool {
	return true
}
