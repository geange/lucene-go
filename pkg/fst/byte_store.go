package fst

import (
	"errors"
	"github.com/geange/lucene-go/pkg/collection"
	"io"
)

// TODO: merge with PagedBytes, except PagedBytes doesn't
// let you read while writing which FST needs

type ByteStore struct {
	blocks collection.ArrayList[[]byte]

	blockSize int64
	blockBits int64
	blockMask int64

	current []byte

	nextWrite int64
}

// WriteByteAt Absolute write byte; you must ensure dest is < max position written so far.
func (r *ByteStore) WriteByteAt(dest int64, b byte) error {
	blockIndex := dest >> r.blockBits
	block, err := r.blocks.Get(blockIndex)
	if err != nil {
		return err
	}
	block[dest&r.blockMask] = b
	return nil
}

func (r *ByteStore) WriteByte(b byte) error {
	if r.nextWrite == r.blockSize {
		r.current = make([]byte, r.blockSize)
		err := r.blocks.Add(r.current)
		if err != nil {
			return err
		}
		r.nextWrite = 0
	}
	r.current[r.nextWrite] = b
	r.nextWrite++
	return nil
}

func (r *ByteStore) WriteBytes(bs []byte) error {
	offset := int64(0)
	size := int64(len(bs))

	for size > 0 {
		chunk := r.blockSize - r.nextWrite
		if size <= chunk {
			copy(r.current[r.nextWrite:], bs[offset:])
			r.nextWrite += size
			break
		} else {
			if chunk > 0 {
				copy(r.current[r.nextWrite:], bs[offset:offset+chunk])
				offset += chunk
				size -= chunk
			}
			r.current = make([]byte, r.blockSize)
			if err := r.blocks.Add(r.current); err != nil {
				return err
			}
			r.nextWrite = 0
		}
	}
	return nil
}

func (r *ByteStore) GetBlockBits() int64 {
	return r.blockBits
}

// WriteBytesAt Absolute writeBytes without changing the current position.
// Note: this cannot "grow" the bytes, so you must only call it on already written parts.
func (r *ByteStore) WriteBytesAt(dest int64, bs []byte) error {
	size := int64(len(bs))

	end := dest + int64(len(bs))
	blockIndex := end >> r.blockBits
	downTo := end & r.blockMask
	if downTo == 0 {
		blockIndex--
		downTo = r.blockSize
	}

	block, err := r.blocks.Get(int(blockIndex))
	if err != nil {
		return err
	}

	for size > 0 {
		if size <= downTo {
			copy(block[downTo-size:downTo], bs)
			break
		}

		size -= downTo
		copy(block[0:downTo], bs[size:])
		blockIndex--

		block, err = r.blocks.Get(int(blockIndex))
		if err != nil {
			return err
		}
		downTo = r.blockSize
	}
	return nil
}

/**

  long end = src + len;

  int blockIndex = (int) (end >> blockBits);
  int downTo = (int) (end & blockMask);
  if (downTo == 0) {
    blockIndex--;
    downTo = blockSize;
  }
  byte[] block = blocks.get(blockIndex);

  while (len > 0) {
    //System.out.println("  cycle downTo=" + downTo);
    if (len <= downTo) {
      //System.out.println("    finish");
      writeBytes(dest, block, downTo-len, len);
      break;
    } else {
      //System.out.println("    partial");
      len -= downTo;
      writeBytes(dest + len, block, 0, downTo);
      blockIndex--;
      block = blocks.get(blockIndex);
      downTo = blockSize;
    }
  }

*/

// CopyBytes Absolute copy bytes self to self, without changing the position.
// Note: this cannot "grow" the bytes, so must only call it on already written parts.
func (r *ByteStore) CopyBytes(src, dest, size int64) error {
	if src >= dest {
		return errors.New("src >= dest")
	}

	end := src + size
	blockIndex := end >> r.blockBits
	downTo := end & r.blockMask
	if downTo == 0 {
		blockIndex--
		downTo = r.blockSize
	}

	block, err := r.blocks.Get(int(blockIndex))
	if err != nil {
		return err
	}

	for size > 0 {
		if size <= downTo {
			err := r.WriteBytesAt(dest, block[downTo-size:size])
			if err != nil {
				return err
			}
			break
		}

		size -= downTo
		err := r.WriteBytesAt(dest+size, block[0:downTo])
		if err != nil {
			return err
		}
		blockIndex--
		block, err = r.blocks.Get(int(blockIndex))
		if err != nil {
			return err
		}
		downTo = r.blockSize
	}
	return nil
}

/**

  public void copyBytes(long src, byte[] dest, int offset, int len) {
    int blockIndex = (int) (src >> blockBits);
    int upto = (int) (src & blockMask);
    byte[] block = blocks.get(blockIndex);
    while (len > 0) {
      int chunk = blockSize - upto;
      if (len <= chunk) {
        System.arraycopy(block, upto, dest, offset, len);
        break;
      } else {
        System.arraycopy(block, upto, dest, offset, chunk);
        blockIndex++;
        block = blocks.get(blockIndex);
        upto = 0;
        len -= chunk;
        offset += chunk;
      }
    }
  }


*/

// CopyBytesToArray Copies bytes from this store to a target byte array.
func (r *ByteStore) CopyBytesToArray(src int64, dest []byte) error {
	blockIndex := src >> r.blockBits
	upto := src & r.blockMask

	block, err := r.blocks.Get(int(blockIndex))
	if err != nil {
		return err
	}

	offset, size := int64(0), int64(len(dest))

	for size > 0 {
		chunk := r.blockSize - upto
		if size <= chunk {
			copy(dest[offset:offset+size], block[upto:])
			break
		}

		copy(dest[offset:offset+chunk], block[upto:])
		blockIndex++

		block, err = r.blocks.Get(int(blockIndex))
		if err != nil {
			return err
		}

		upto = 0
		size -= chunk
		offset += chunk
	}
	return nil
}

/**
  public void writeInt(long pos, int value) {
    int blockIndex = (int) (pos >> blockBits);
    int upto = (int) (pos & blockMask);
    byte[] block = blocks.get(blockIndex);
    int shift = 24;
    for(int i=0;i<4;i++) {
      block[upto++] = (byte) (value >> shift);
      shift -= 8;
      if (upto == blockSize) {
        upto = 0;
        blockIndex++;
        block = blocks.get(blockIndex);
      }
    }
  }

*/

// WriteInt32 Writes an int at the absolute position without changing the current pointer.
func (r *ByteStore) WriteInt32(pos int64, value int32) error {
	blockIndex := pos >> r.blockBits
	upto := pos & r.blockMask
	block, err := r.blocks.Get(blockIndex)
	if err != nil {
		return err
	}
	shift := 24

	for i := 0; i < 4; i++ {
		block[upto] = (byte)(value >> shift)
		upto++
		shift -= 8
		if upto == r.blockSize {
			upto = 0
			blockIndex++
			block, err = r.blocks.Get(blockIndex)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Reverse from srcPos, inclusive, to destPos, inclusive.
func (r *ByteStore) Reverse(srcPos, destPos int64) error {
	if srcPos > destPos {
		return errors.New("srcPos > destPos")
	}
	if destPos > r.GetPosition() {
		return errors.New("destPos bigger than position")
	}

	srcBlockIndex := srcPos >> r.blockBits
	src := srcPos & r.blockMask
	srcBlock, err := r.blocks.Get(srcBlockIndex)
	if err != nil {
		return err
	}

	destBlockIndex := destPos >> r.blockBits
	dest := destPos & r.blockMask
	destBlock, err := r.blocks.Get(destBlockIndex)
	if err != nil {
		return err
	}

	limit := (destPos - srcPos + 1) / 2

	for i := int64(0); i < limit; i++ {
		b := srcBlock[src]
		srcBlock[src] = destBlock[dest]
		destBlock[dest] = b
		src++

		if src == r.blockSize {
			srcBlockIndex++
			srcBlock, err = r.blocks.Get(srcBlockIndex)
			if err != nil {
				return err
			}
			src = 0
		}

		dest--

		if dest == -1 {
			destBlockIndex--
			destBlock, err = r.blocks.Get(destBlockIndex)
			if err != nil {
				return err
			}
			dest = r.blockSize - 1
		}
	}
	return nil
}

func (r *ByteStore) SkipBytes(size int64) error {
	for size > 0 {
		chunk := r.blockSize - r.nextWrite
		if size <= chunk {
			r.nextWrite += size
			break
		}

		size -= chunk
		current := make([]byte, r.blockSize)
		err := r.blocks.Add(current)
		if err != nil {
			return err
		}
		r.nextWrite = 0
	}
	return nil
}

func (r *ByteStore) GetPosition() int64 {
	return int64(r.blocks.Size()-1)*r.blockSize + r.nextWrite
}

// Truncate Pos must be less than the max position written so far! Ie, you cannot "grow" the file with this!
func (r *ByteStore) Truncate(newLen int64) error {
	if newLen > r.GetPosition() {
		return errors.New("newLen > r.GetPosition()")
	}

	if newLen < 0 {
		return errors.New("newLen < 0")
	}

	blockIndex := newLen >> r.blockBits
	nextWrite := newLen & r.blockMask
	if nextWrite == 0 {
		blockIndex--
		nextWrite = r.blockSize
	}

	err := r.blocks.ClearSubList(blockIndex+1, r.blocks.Size())
	if err != nil {
		return err
	}

	if newLen == 0 {
		r.current = nil
	} else {
		r.current, err = r.blocks.Get(blockIndex)
		if err != nil {
			return err
		}
	}

	if newLen != r.GetPosition() {
		return errors.New("newLen != r.GetPosition()")
	}

	return nil
}

func (r *ByteStore) Finish() error {
	if r.current != nil {
		// 减少内存消耗
		lastBuffer := make([]byte, r.nextWrite)
		copy(lastBuffer[:r.nextWrite], r.current[:r.nextWrite])
		err := r.blocks.Set(r.blocks.Size()-1, lastBuffer)
		if err != nil {
			return err
		}
		r.current = nil
	}
	return nil
}

// WriteTo Writes all of our bytes to the target DataOutput.
func (r *ByteStore) WriteTo(out io.Writer) error {
	for _, block := range r.blocks.List() {
		_, err := out.Write(block)
		if err != nil {
			return err
		}
	}
	return nil
}
