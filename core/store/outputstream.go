package store

import (
	"bufio"
	"io"
)

var _ IndexOutput = &OutputStream{}

// OutputStream Implementation class for buffered IndexOutput that writes to an OutputStream.
type OutputStream struct {
	*BaseIndexOutput

	out          *bufio.Writer
	closer       io.Closer
	bytesWritten int64
	crc          Hash
}

func (o *OutputStream) GetChecksum() (uint32, error) {
	return o.crc.Sum(), nil
}

func NewOutputStream(name string, out io.WriteCloser) *OutputStream {
	output := &OutputStream{
		out:    bufio.NewWriter(out),
		closer: out,
		crc:    NewHash(),
	}
	output.BaseIndexOutput = NewBaseIndexOutput(name, output)
	return output
}

func (o *OutputStream) Write(b []byte) (int, error) {
	o.crc.Write(b)

	o.bytesWritten += int64(len(b))
	return o.out.Write(b)
}

func (o *OutputStream) Close() error {
	if err := o.out.Flush(); err != nil {
		return err
	}
	return o.closer.Close()
}

func (o *OutputStream) GetFilePointer() int64 {
	return o.bytesWritten
}
