package fst

import "github.com/geange/lucene-go/core/store"

var _ BytesReader = &bytesStoreReverseReader{}

func (bs *BytesStore) newBytesStoreReverseReader() *bytesStoreReverseReader {
	input := &bytesStoreReverseReader{
		nextBuffer: -1,
		nextRead:   bs.blockSize,
		ptr:        bs,
	}

	if bs.blocks.Size() != 0 {
		block, err := bs.blocks.Get(0)
		if err != nil {
			return nil
		}
		input.current = block
	}

	input.DataInputImp = store.NewDataInputImp(input)

	return input
}

type bytesStoreReverseReader struct {
	*store.DataInputImp

	current    []byte
	nextBuffer int
	nextRead   int
	ptr        *BytesStore
}

/**
  public byte readByte() {
    if (nextRead == -1) {
      current = blocks.get(nextBuffer--);
      nextRead = blockSize-1;
    }
    return current[nextRead--];
  }
*/

func (r *bytesStoreReverseReader) ReadByte() (byte, error) {
	if r.nextRead == -1 {
		block, err := r.ptr.blocks.Get(r.nextBuffer)
		if err != nil {
			return 0, err
		}
		r.nextBuffer++
		r.current = block
		r.nextRead = r.ptr.blockSize - 1
	}
	b := r.current[r.nextRead]
	r.nextRead--
	return b, nil
}

/**

  public void readBytes(byte[] b, int offset, int len) {
    for(int i=0;i<len;i++) {
      b[offset+i] = readByte();
    }
  }

*/

func (r *bytesStoreReverseReader) ReadBytes(b []byte) error {
	for i := range b {
		v, err := r.ReadByte()
		if err != nil {
			return err
		}
		b[i] = v
	}
	return nil
}

/**

  public void skipBytes(long count) {
    setPosition(getPosition() - count);
  }


*/

func (r *bytesStoreReverseReader) SkipBytes(count int) error {
	r.SetPosition(r.GetPosition() - count)
	return nil
}

func (r *bytesStoreReverseReader) GetPosition() int {
	return (r.nextBuffer+1)*r.ptr.blockSize + r.nextRead
}

/**

  public void setPosition(long pos) {
    // NOTE: a little weird because if you
    // setPosition(0), the next byte you read is
    // bytes[0] ... but I would expect bytes[-1] (ie,
    // EOF)...?
    int bufferIndex = (int) (pos >> blockBits);
    if (nextBuffer != bufferIndex - 1) {
      nextBuffer = bufferIndex - 1;
      current = blocks.get(bufferIndex);
    }
    nextRead = (int) (pos & blockMask);
    assert getPosition() == pos : "pos=" + pos + " getPos()=" + getPosition();
  }


*/

func (r *bytesStoreReverseReader) SetPosition(pos int) {
	// NOTE: a little weird because if you
	// setPosition(0), the next byte you read is
	// bytes[0] ... but I would expect bytes[-1] (ie,
	// EOF)...?
	bufferIndex := pos >> r.ptr.blockBits
	if r.nextBuffer != bufferIndex-1 {
		r.nextBuffer = bufferIndex - 1
		block, err := r.ptr.blocks.Get(bufferIndex)
		if err != nil {
			return
		}
		r.current = block
	}
	r.nextRead = pos & r.ptr.blockMask
}

func (r *bytesStoreReverseReader) Reversed() bool {
	return true
}
