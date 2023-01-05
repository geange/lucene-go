package simpletext

import (
	"bytes"
	"fmt"
	"github.com/geange/lucene-go/core/store"
)

var (
	NEWLINE  = byte('\n')
	ESCAPE   = byte('\\')
	CHECKSUM = []byte("checksum ")
)

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
