package index

import (
	"bytes"
	"github.com/geange/lucene-go/core/store"
)

// PagedBytes Represents a logical byte[] as a series of pages.
// You can write-once into the logical byte[] (append only), using copy,
// and then retrieve slices (BytesRef) into it using fill.
// lucene.internal
// TODO: refactor this, byteblockpool, fst.bytestore, and any
// other "shift/mask big arrays". there are too many of these classes!
type PagedBytes struct {
	blocks            [][]byte
	numBlocks         int
	blockSize         int
	blockBits         int
	blockMask         int
	didSkipBytes      bool
	frozen            bool
	upto              int
	currentBlock      []byte
	bytesUsedPerBlock int64
}

func (r *PagedBytes) GetPointer() int64 {
	if len(r.currentBlock) == 0 {
		return 0
	} else {
		return int64((r.numBlocks * r.blockSize) + r.upto)
	}
}

func (r *PagedBytes) addBlock(block []byte) {
	r.numBlocks++
	r.blocks = append(r.blocks, block)
}

// CopyV1 Read this many bytes from in
func (r *PagedBytes) CopyV1(in store.IndexInput, byteCount int) error {
	for byteCount > 0 {
		left := r.blockSize - r.upto
		if left == 0 {
			if len(r.currentBlock) != 0 {
				r.addBlock(r.currentBlock)
			}
			r.currentBlock = make([]byte, r.blockSize)
			r.upto = 0
			left = r.blockSize
		}
		if left < byteCount {
			if _, err := in.Read(r.currentBlock[r.upto : r.upto+left]); err != nil {
				return err
			}
			r.upto = r.blockSize
			byteCount -= left
		} else {
			if _, err := in.Read(r.currentBlock[r.upto : r.upto+byteCount]); err != nil {
				return err
			}
			r.upto += byteCount
			break
		}
	}
	return nil
}

// CopyV2 Copy BytesRef in, setting BytesRef out to the result.
// Do not use this if you will use freeze(true). This only supports bytes.length <= blockSize
func (r *PagedBytes) CopyV2(bytes []byte, out *bytes.Buffer) error {
	left := r.blockSize - r.upto
	if len(bytes) > left || len(r.currentBlock) == 0 {
		if len(r.currentBlock) != 0 {
			r.addBlock(r.currentBlock)
			r.didSkipBytes = true
		}
		r.currentBlock = make([]byte, r.blockSize)
		r.upto = 0
		left = r.blockSize
		//assert bytes.length <= blockSize;
		// TODO: we could also support variable block sizes
	}

	copy(r.currentBlock[r.upto:], out.Bytes())
	r.upto += len(bytes)
	return nil
}

// PagedBytesReader Provides methods to read BytesRefs from a frozen PagedBytes.
type PagedBytesReader struct {
}

var _ store.DataInput = &PagedBytesDataInput{}

type PagedBytesDataInput struct {
	*store.DataInputDefault
	*PagedBytes

	currentBlockIndex int
	currentBlockUpto  int
	currentBlock      []byte
}

func NewPagedBytesDataInput(pageBytes *PagedBytes) *PagedBytesDataInput {
	input := &PagedBytesDataInput{
		PagedBytes:   pageBytes,
		currentBlock: pageBytes.blocks[0],
	}
	input.DataInputDefault = store.NewDataInputDefault(&store.DataInputDefaultConfig{
		ReadByte: input.ReadByte,
		Read:     input.Read,
	})
	return input
}

func (r *PagedBytesDataInput) ReadByte() (byte, error) {
	if r.currentBlockUpto == r.blockSize {
		r.nextBlock()
	}
	value := r.currentBlock[r.currentBlockUpto]
	r.currentBlockUpto++
	return value, nil
}

func (r *PagedBytesDataInput) Read(bs []byte) (n int, err error) {
	offset := 0
	offsetEnd := len(bs)

	for {
		blockLeft := r.blockSize - r.currentBlockUpto
		left := offsetEnd - offsetEnd
		if blockLeft < left {
			copy(bs[offset:], r.currentBlock[r.currentBlockUpto:r.currentBlockUpto+blockLeft])
			r.nextBlock()
			offset += blockLeft
		} else {
			copy(bs[offset:], r.currentBlock[r.currentBlockUpto:r.currentBlockUpto+left])
			r.currentBlockUpto += left
			break
		}
	}
	return len(bs), nil
}

func (r *PagedBytesDataInput) nextBlock() {
	r.currentBlockIndex++
	r.currentBlockUpto = 0
	r.currentBlock = r.blocks[r.currentBlockIndex]
}

// Returns the current byte position.
func (r *PagedBytesDataInput) getPosition() int64 {
	return int64(r.currentBlockIndex*r.blockSize + r.currentBlockUpto)
}

var _ store.DataOutput = &PagedBytesDataOutput{}

type PagedBytesDataOutput struct {
	*store.DataOutputDefault
	*PagedBytes
}

func (r *PagedBytesDataOutput) WriteByte(b byte) error {
	if r.upto == r.blockSize {
		if len(r.currentBlock) != 0 {
			r.addBlock(r.currentBlock)
		}
		r.currentBlock = make([]byte, r.blockBits)
		r.upto = 0
	}
	r.currentBlock[r.upto] = b
	r.upto++
	return nil
}

func (r *PagedBytesDataOutput) Write(bs []byte) (n int, err error) {
	offset, length := 0, len(bs)

	if length == 0 {
		return
	}

	if r.upto == r.blockSize {
		if len(r.currentBlock) != 0 {
			r.addBlock(r.currentBlock)
		}
		r.currentBlock = make([]byte, r.blockSize)
		r.upto = 0
	}

	offsetEnd := len(bs)
	for {
		left := offsetEnd - offset
		blockLeft := r.blockSize - r.upto
		if blockLeft < left {
			copy(r.currentBlock[r.upto:], bs[offset:offset+blockLeft])
			r.addBlock(r.currentBlock)
			r.currentBlock = make([]byte, r.blockSize)
			r.upto = 0
			offset += blockLeft
		} else {
			// Last block
			copy(r.currentBlock[r.upto:], bs[offset:offset+left])
			r.upto += left
			break
		}
	}
	return len(bs), nil
}

// GetPosition Return the current byte position.
func (r *PagedBytesDataOutput) GetPosition() int64 {
	return r.GetPointer()
}
