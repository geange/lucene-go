package store

import (
	"bufio"
	"hash"
	"hash/crc32"
	"io"
)

var _ IndexOutput = &OutputStreamIndexOutput{}

// OutputStreamIndexOutput Implementation class for buffered IndexOutput that writes to an OutputStream.
type OutputStreamIndexOutput struct {
	*IndexOutputDefault
	out    *bufio.Writer
	closer io.Closer

	bytesWritten   int64
	flushedOnClose bool
	crc            hash.Hash32
}

func (o *OutputStreamIndexOutput) GetChecksum() (uint32, error) {
	if err := o.out.Flush(); err != nil {
		return 0, err
	}
	return o.crc.Sum32(), nil
}

func NewOutputStreamIndexOutput(name string, out io.WriteCloser) *OutputStreamIndexOutput {
	output := &OutputStreamIndexOutput{
		out:    bufio.NewWriter(out),
		closer: out,
		crc:    crc32.NewIEEE(),
	}
	output.IndexOutputDefault = NewIndexOutputDefault(name, output)
	return output
}

func (o *OutputStreamIndexOutput) WriteByte(b byte) error {
	if _, err := o.crc.Write([]byte{b}); err != nil {
		return err
	}

	o.bytesWritten++
	_, err := o.out.Write([]byte{b})
	return err
}

func (o *OutputStreamIndexOutput) Write(b []byte) (int, error) {
	if _, err := o.crc.Write(b); err != nil {
		return 0, err
	}

	o.bytesWritten += int64(len(b))
	return o.out.Write(b)
}

func (o *OutputStreamIndexOutput) Close() error {
	if err := o.out.Flush(); err != nil {
		return err
	}
	return o.closer.Close()
}

func (o *OutputStreamIndexOutput) GetFilePointer() int64 {
	return o.bytesWritten
}
