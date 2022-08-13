package store

import (
	"bufio"
	"hash"
	"io"
)

var _ IndexOutput = &OutputStreamIndexOutput{}

// OutputStreamIndexOutput Implementation class for buffered IndexOutput that writes to an OutputStream.
type OutputStreamIndexOutput struct {
	*DataOutputImp
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

func NewOutputStreamIndexOutput(out io.WriteCloser) *OutputStreamIndexOutput {
	output := &OutputStreamIndexOutput{
		out:    bufio.NewWriter(out),
		closer: out,
	}
	output.DataOutputImp = NewDataOutputImp(output)
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

func (o *OutputStreamIndexOutput) WriteBytes(b []byte) error {
	if _, err := o.crc.Write(b); err != nil {
		return err
	}

	o.bytesWritten += int64(len(b))
	_, err := o.out.Write(b)
	return err
}

func (o *OutputStreamIndexOutput) Close() error {
	return o.closer.Close()
}

func (o *OutputStreamIndexOutput) GetFilePointer() int64 {
	return o.bytesWritten
}
