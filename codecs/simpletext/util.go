package simpletext

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	"github.com/geange/lucene-go/core/store"
)

var (
	NEWLINE  = byte('\n')
	ESCAPE   = byte('\\')
	CHECKSUM = []byte("checksum ")
)

type TextReader struct {
	in  store.IndexInput
	buf *bytes.Buffer
}

func NewTextReader(in store.IndexInput, buf *bytes.Buffer) *TextReader {
	return &TextReader{
		in:  in,
		buf: buf,
	}
}

func (t *TextReader) ReadLabel(label []byte) (string, error) {
	t.buf.Reset()
	if err := ReadLine(t.in, t.buf); err != nil {
		return "", err
	}

	if !bytes.HasPrefix(t.buf.Bytes(), label) {
		return "", fmt.Errorf("label not found:%s", string(label))
	}

	t.buf.Next(len(label))
	return t.buf.String(), nil
}

func (t *TextReader) StartsWith(label []byte) (bool, error) {
	t.buf.Reset()
	if err := ReadLine(t.in, t.buf); err != nil {
		return false, err
	}
	return bytes.HasPrefix(t.buf.Bytes(), label), nil
}

func (t *TextReader) ParseInt(prefix []byte) (int, error) {
	v, err := t.ReadLabel(prefix)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(v)
}

type TextWriter struct {
	out store.IndexOutput
}

func (t *TextWriter) Write(v any) error {
	switch v.(type) {
	case []byte:
		return WriteBytes(t.out, v.([]byte))
	case string:
		return WriteString(t.out, v.(string))
	default:
		return errors.New("unsupported type")
	}
}

func readValue(out store.IndexInput, label []byte, buf *bytes.Buffer) (string, error) {
	buf.Reset()
	if err := ReadLine(out, buf); err != nil {
		return "", err
	}

	if !bytes.HasPrefix(buf.Bytes(), label) {
		return "", fmt.Errorf("label not found:%s", string(label))
	}
	buf.Next(len(label))
	//
	//buf.Truncate(len(label))

	return buf.String(), nil
}

func WriteString(out store.DataOutput, s string) error {
	return WriteBytes(out, []byte(s))
}

func WriteBytes(out store.DataOutput, bs []byte) error {
	for i := range bs {
		if bs[i] == NEWLINE || bs[i] == ESCAPE {
			if err := out.WriteByte(ESCAPE); err != nil {
				return err
			}
			if err := out.WriteByte(bs[i]); err != nil {
				return err
			}
		}
		if err := out.WriteByte(bs[i]); err != nil {
			return err
		}
	}
	return nil
}

func WriteNewline(out store.DataOutput) error {
	return out.WriteByte(NEWLINE)
}

func ReadLine(in store.IndexInput, buf *bytes.Buffer) error {
	buf.Reset()

	for {
		b, err := in.ReadByte()
		if err != nil {
			return err
		}
		if b == ESCAPE {
			b, err = in.ReadByte()
			if err != nil {
				return err
			}
			buf.WriteByte(b)
		} else {
			if b == NEWLINE {
				break
			}
			buf.WriteByte(b)
		}
	}

	return nil
}

func WriteChecksum(out store.IndexOutput) error {
	checksum, err := out.GetChecksum()
	if err != nil {
		return err
	}

	if err := WriteBytes(out, CHECKSUM); err != nil {
		return err
	}
	if err != nil {
		return err
	}
	if err := WriteString(out, fmt.Sprintf("%020d", checksum)); err != nil {
		return err
	}
	return WriteNewline(out)
}

func CheckFooter(input store.ChecksumIndexInput) error {
	scratch := new(bytes.Buffer)

	checksum := input.GetChecksum()

	if err := ReadLine(input, scratch); err != nil {
		return err
	}

	line := scratch.Bytes()

	if !bytes.HasPrefix(line, CHECKSUM) {
		return fmt.Errorf("simpleText failure: expected checksum line but got (%s)", string(line))
	}

	expectedChecksum := []byte(fmt.Sprintf("%020d", checksum))
	actualChecksum := line[len(CHECKSUM):]
	if !bytes.Equal(expectedChecksum, actualChecksum) {
		return fmt.Errorf("simpleText checksum failure: (%s) != (%s)", expectedChecksum, actualChecksum)
	}

	if input.Length() != input.GetFilePointer() {
		return fmt.Errorf("unexpected stuff at the end of file, please be careful with your text editor")
	}
	return nil
}

func ParseInt(scratch *bytes.Buffer, prefix []byte) (int, error) {
	if !bytes.HasPrefix(scratch.Bytes(), prefix) {
		return 0, fmt.Errorf("prefix is not %s", string(prefix))
	}
	scratch.Next(len(prefix))
	return strconv.Atoi(scratch.String())
}
