package fst

import (
	"encoding/binary"
	"github.com/geange/lucene-go/core/store"
	. "github.com/geange/lucene-go/math"
	"github.com/geange/lucene-go/util/structure"
)

var _ store.DataOutput = &BytesStore{}

// BytesStore TODO: merge with PagedBytes, except PagedBytes doesn't
// let you read while writing which FST needs
type BytesStore struct {
	*store.DataOutputImp

	blocks *structure.ArrayList[[]byte]

	blockSize int
	blockBits int
	blockMask int
	current   []byte
	nextWrite int
}

func NewBytesStore(blockBits int) *BytesStore {
	blockSize := 1 << blockBits

	input := &BytesStore{
		blocks:    structure.NewArrayList[[]byte](),
		blockSize: blockSize,
		blockBits: blockBits,
		blockMask: blockSize - 1,
		nextWrite: blockSize,
	}
	input.DataOutputImp = store.NewDataOutputImp(input)
	return input
}

/**

  public BytesStore(DataInput in, long numBytes, int maxBlockSize) throws IOException {
    int blockSize = 2;
    int blockBits = 1;
    while(blockSize < numBytes && blockSize < maxBlockSize) {
      blockSize *= 2;
      blockBits++;
    }
    this.blockBits = blockBits;
    this.blockSize = blockSize;
    this.blockMask = blockSize-1;
    long left = numBytes;
    while(left > 0) {
      final int chunk = (int) Math.min(blockSize, left);
      byte[] block = new byte[chunk];
      in.readBytes(block, 0, block.length);
      blocks.add(block);
      left -= chunk;
    }

    // So .getPosition still works
    nextWrite = blocks.get(blocks.size()-1).length;
  }


*/

// NewBytesStoreFromDataInput Pulls bytes from the provided IndexInput.
func NewBytesStoreFromDataInput(in store.DataInput, numBytes, maxBlockSize int) (*BytesStore, error) {
	blockSize, blockBits := 2, 1

	for blockSize < numBytes && blockSize < maxBlockSize {
		blockSize *= 2
		blockBits++
	}

	input := &BytesStore{
		blocks:    structure.NewArrayList[[]byte](),
		blockSize: blockSize,
		blockBits: blockBits,
		blockMask: blockSize - 1,
	}
	input.DataOutputImp = store.NewDataOutputImp(input)

	left := numBytes
	for left > 0 {
		chunk := Min(blockSize, left)
		block := make([]byte, chunk)
		err := in.ReadBytes(block)
		if err != nil {
			return nil, err
		}
		input.blocks.Add(block)
		left -= chunk
	}

	// So .getPosition still works
	idx := input.blocks.Size() - 1
	last, err := input.blocks.Get(idx)
	if err != nil {
		return nil, err
	}
	input.nextWrite = len(last)

	return input, nil
}

/**
  public void writeByte(long dest, byte b) {
    int blockIndex = (int) (dest >> blockBits);
    byte[] block = blocks.get(blockIndex);
    block[(int) (dest & blockMask)] = b;
  }

*/

// WriteByteAt Absolute write byte; you must ensure dest is < max position written so far.
func (bs *BytesStore) WriteByteAt(dest int, s byte) error {
	blockIndex := dest >> bs.blockBits
	block, err := bs.blocks.Get(blockIndex)
	if err != nil {
		return err
	}

	idx := dest & bs.blockMask
	if idx >= len(block) {
		return ErrOutOfArrayRange
	}
	block[idx] = s
	return nil
}

/**
  public void writeByte(byte b) {
    if (nextWrite == blockSize) {
      current = new byte[blockSize];
      blocks.add(current);
      nextWrite = 0;
    }
    current[nextWrite++] = b;
  }

*/

func (bs *BytesStore) WriteByte(b byte) error {
	if bs.nextWrite == bs.blockSize {
		bs.current = make([]byte, bs.blockSize)
		bs.blocks.Add(bs.current)
		bs.nextWrite = 0
	}
	bs.current[bs.nextWrite] = b
	bs.nextWrite++
	return nil
}

/**
  public void writeBytes(byte[] b, int offset, int len) {
    while (len > 0) {
      int chunk = blockSize - nextWrite;
      if (len <= chunk) {
        assert b != null;
        assert current != null;
        System.arraycopy(b, offset, current, nextWrite, len);
        nextWrite += len;
        break;
      } else {
        if (chunk > 0) {
          System.arraycopy(b, offset, current, nextWrite, chunk);
          offset += chunk;
          len -= chunk;
        }
        current = new byte[blockSize];
        blocks.add(current);
        nextWrite = 0;
      }
    }
  }

*/

func (bs *BytesStore) WriteBytes(b []byte) error {
	offset, size := 0, len(b)

	for size > 0 {
		chunk := bs.blockSize - bs.nextWrite
		if size <= chunk {
			copy(bs.current[bs.nextWrite:], b)
			bs.nextWrite += size
			break
		} else {
			if chunk > 0 {
				copy(bs.current[bs.nextWrite:], b[offset:offset+chunk])
				offset += chunk
				size -= chunk
			}

			bs.current = make([]byte, bs.blockSize)
			bs.blocks.Add(bs.current)
			bs.nextWrite = 0
		}
	}
	return nil
}

func (bs *BytesStore) getBlockBits() int {
	return bs.blockSize
}

/**
  void writeBytes(long dest, byte[] b, int offset, int len) {
    //System.out.println("  BS.writeBytes dest=" + dest + " offset=" + offset + " len=" + len);
    assert dest + len <= getPosition(): "dest=" + dest + " pos=" + getPosition() + " len=" + len;

    // Note: weird: must go "backwards" because copyBytes
    // calls us with overlapping src/dest.  If we
    // go forwards then we overwrite bytes before we can
    // copy them:

    final long end = dest + len;
    int blockIndex = (int) (end >> blockBits);
    int downTo = (int) (end & blockMask);
    if (downTo == 0) {
      blockIndex--;
      downTo = blockSize;
    }
    byte[] block = blocks.get(blockIndex);

    while (len > 0) {
      //System.out.println("    cycle downTo=" + downTo + " len=" + len);
      if (len <= downTo) {
        //System.out.println("      final: offset=" + offset + " len=" + len + " dest=" + (downTo-len));
        System.arraycopy(b, offset, block, downTo-len, len);
        break;
      } else {
        len -= downTo;
        //System.out.println("      partial: offset=" + (offset + len) + " len=" + downTo + " dest=0");
        System.arraycopy(b, offset + len, block, 0, downTo);
        blockIndex--;
        block = blocks.get(blockIndex);
        downTo = blockSize;
      }
    }
  }

*/

// WriteBytesAt Absolute writeBytes without changing the current position. Note: this cannot "grow" the bytes,
// so you must only call it on already written parts.
func (bs *BytesStore) WriteBytesAt(dest int, b []byte) error {
	// Note: weird: must go "backwards" because copyBytes
	// calls us with overlapping src/dest.  If we
	// go forwards then we overwrite bytes before we can
	// copy them:

	offset, size := 0, len(b)

	end := dest + size
	blockIndex := end >> bs.blockSize
	downTo := end & bs.blockMask
	if downTo == 0 {
		blockIndex--
		downTo = bs.blockSize
	}

	block, err := bs.blocks.Get(blockIndex)
	if err != nil {
		return err
	}

	for size > 0 {
		if size <= downTo {
			copy(block[downTo-size:], b)
			break
		} else {
			size -= downTo
			copy(block[:downTo], b[offset+size:])
			blockIndex--
			block, _ = bs.blocks.Get(blockIndex)
			downTo = bs.blockSize
		}
	}
	return nil
}

/**

  public void copyBytes(long src, long dest, int len) {
    //System.out.println("BS.copyBytes src=" + src + " dest=" + dest + " len=" + len);
    assert src < dest;

    // Note: weird: must go "backwards" because copyBytes
    // calls us with overlapping src/dest.  If we
    // go forwards then we overwrite bytes before we can
    // copy them:

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
  }

*/

// CopyBytesSelf Absolute copy bytes self to self, without changing the position. Note: this cannot "grow" the bytes, so must only call it on already written parts.
func (bs *BytesStore) CopyBytesSelf(src, dest, size int) error {
	// Note: weird: must go "backwards" because copyBytes
	// calls us with overlapping src/dest.  If we
	// go forwards then we overwrite bytes before we can
	// copy them:

	end := src + size
	blockIndex := end >> bs.blockSize
	downTo := end & bs.blockMask
	if downTo == 0 {
		blockIndex--
		downTo = bs.blockSize
	}

	block, err := bs.blocks.Get(blockIndex)
	if err != nil {
		return err
	}

	for size > 0 {
		if size <= downTo {
			err := bs.WriteBytesAt(dest, block[downTo-size:downTo])
			if err != nil {
				return err
			}
			break
		} else {
			size -= downTo
			err := bs.WriteBytesAt(dest+size, block[0:downTo])
			if err != nil {
				return err
			}
			blockIndex--
			block, err = bs.blocks.Get(blockIndex)
			if err != nil {
				return err
			}
			downTo = bs.blockSize
		}
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
func (bs *BytesStore) CopyBytesToArray(src int, dest []byte) error {
	blockIndex := src >> bs.blockBits
	upto := src & bs.blockMask
	block, err := bs.blocks.Get(blockIndex)
	if err != nil {
		return err
	}

	offset, size := 0, len(dest)
	for size > 0 {
		chunk := bs.blockSize - upto
		if size <= chunk {
			copy(dest, block[upto:upto+size])
			break
		} else {
			copy(dest[offset:], block[upto:upto+chunk])
			blockIndex++
			block, err = bs.blocks.Get(blockIndex)
			if err != nil {
				return err
			}
			upto = 0
			size -= chunk
			offset += chunk
		}
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

// WriteInt Writes an int at the absolute position without changing the current pointer.
func (bs *BytesStore) WriteInt(pos, value int) error {
	buff := make([]byte, 4)
	binary.BigEndian.PutUint32(buff, uint32(value))
	return bs.WriteBytesAt(pos, buff)
}

/**
  public void reverse(long srcPos, long destPos) {

    int srcBlockIndex = (int) (srcPos >> blockBits);
    int src = (int) (srcPos & blockMask);
    byte[] srcBlock = blocks.get(srcBlockIndex);

    int destBlockIndex = (int) (destPos >> blockBits);
    int dest = (int) (destPos & blockMask);
    byte[] destBlock = blocks.get(destBlockIndex);

    int limit = (int) (destPos - srcPos + 1)/2;
    for(int i=0;i<limit;i++) {
      byte b = srcBlock[src];
      srcBlock[src] = destBlock[dest];
      destBlock[dest] = b;
      src++;
      if (src == blockSize) {
        srcBlockIndex++;
        srcBlock = blocks.get(srcBlockIndex);
        src = 0;
      }

      dest--;
      if (dest == -1) {
        destBlockIndex--;
        destBlock = blocks.get(destBlockIndex);
        dest = blockSize-1;
      }
    }
  }
*/

// Reverse from srcPos, inclusive, to destPos, inclusive.
func (bs *BytesStore) Reverse(srcPos, destPos int) error {

	srcBlockIndex := srcPos >> bs.blockBits
	src := srcPos & bs.blockMask
	srcBlock, err := bs.blocks.Get(srcBlockIndex)
	if err != nil {
		return err
	}

	destBlockIndex := destPos >> bs.blockBits
	dest := destPos & bs.blockMask
	destBlock, err := bs.blocks.Get(destBlockIndex)
	if err != nil {
		return err
	}

	limit := (destPos - srcPos + 1) / 2
	for i := 0; i < limit; i++ {
		srcBlock[src], destBlock[dest] = destBlock[dest], srcBlock[src]

		src++
		if src == bs.blockSize {
			srcBlockIndex++
			srcBlock, err = bs.blocks.Get(srcBlockIndex)
			if err != nil {
				return err
			}
			src = 0
		}

		dest--
		if dest == -1 {
			destBlockIndex--
			destBlock, err = bs.blocks.Get(destBlockIndex)
			if err != nil {
				return err
			}
			dest = bs.blockSize - 1
		}
	}
	return nil
}

/**
  public void skipBytes(int len) {
    while (len > 0) {
      int chunk = blockSize - nextWrite;
      if (len <= chunk) {
        nextWrite += len;
        break;
      } else {
        len -= chunk;
        current = new byte[blockSize];
        blocks.add(current);
        nextWrite = 0;
      }
    }
  }

*/

func (bs *BytesStore) SkipBytes(size int) error {
	for size > 0 {
		chunk := bs.blockSize - bs.nextWrite
		if size <= chunk {
			bs.nextWrite += size
			break
		} else {
			size -= chunk
			bs.current = make([]byte, bs.blockSize)
			bs.blocks.Add(bs.current)
			bs.nextWrite = 0
		}
	}
	return nil
}

/**
  public long getPosition() {
    return ((long) blocks.size()-1) * blockSize + nextWrite;
  }
*/

func (bs *BytesStore) getPosition() int {
	return (bs.blocks.Size()-1)*bs.blockSize + bs.nextWrite
}

/**

  public void truncate(long newLen) {
    assert newLen <= getPosition();
    assert newLen >= 0;
    int blockIndex = (int) (newLen >> blockBits);
    nextWrite = (int) (newLen & blockMask);
    if (nextWrite == 0) {
      blockIndex--;
      nextWrite = blockSize;
    }
    blocks.subList(blockIndex+1, blocks.size()).clear();
    if (newLen == 0) {
      current = null;
    } else {
      current = blocks.get(blockIndex);
    }
    assert newLen == getPosition();
  }


*/

// Truncate Pos must be less than the max position written so far! Ie, you cannot "grow" the file with this!
func (bs *BytesStore) Truncate(newLen int) error {
	blockIndex := newLen >> bs.blockBits
	bs.nextWrite = newLen & bs.blockBits
	if bs.nextWrite == 0 {
		blockIndex--
		bs.nextWrite = bs.blockSize
	}

	err := bs.blocks.Clear(blockIndex+1, bs.blocks.Size())
	if err != nil {
		return err
	}

	if newLen == 0 {
		bs.current = nil
	} else {
		bs.current, err = bs.blocks.Get(blockIndex)
		if err != nil {
			return err
		}
	}
	return nil
}

/**

  public void finish() {
    if (current != null) {
      byte[] lastBuffer = new byte[nextWrite];
      System.arraycopy(current, 0, lastBuffer, 0, nextWrite);
      blocks.set(blocks.size()-1, lastBuffer);
      current = null;
    }
  }

*/

func (bs *BytesStore) Finish() {
	if bs.current != nil {
		lastBuffer := make([]byte, bs.nextWrite)
		copy(lastBuffer, bs.current[0:bs.nextWrite])
		bs.current = nil
	}
}

/**

  public void writeTo(DataOutput out) throws IOException {
    for(byte[] block : blocks) {
      out.writeBytes(block, 0, block.length);
    }
  }


*/

// WriteTo Writes all of our bytes to the target DataOutput.
func (bs *BytesStore) WriteTo(out store.DataOutput) error {
	for _, block := range bs.blocks.Values() {
		err := out.WriteBytes(block)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bs *BytesStore) GetForwardReader() BytesReader {
	if bs.blocks.Size() == 1 {
		block, _ := bs.blocks.Get(0)
		return NewForwardBytesReader(block)
	}
	return bs.newBytesStoreForwardReader()
}

func (bs *BytesStore) GetReverseReader() BytesReader {
	return bs.getReverseReader(true)
}

func (bs *BytesStore) getReverseReader(allowSingle bool) BytesReader {
	if allowSingle && bs.blocks.Size() == 1 {
		block, _ := bs.blocks.Get(0)
		return NewReverseBytesReader(block)
	}
	return bs.newBytesStoreReverseReader()
}
