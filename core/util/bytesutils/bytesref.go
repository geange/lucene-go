package bytesutils

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

var (
	EMPTY_BYTES []byte
)

// BytesRef Represents byte[], as a slice (offset + length) into an existing byte[]. The bytes member should never
// be null; use EMPTY_BYTES if necessary.
// Important note: Unless otherwise noted, Lucene uses this class to represent terms that are encoded as UTF8 bytes
// in the index. To convert them to a Java String (which is UTF16), use utf8ToString. Using code like
// new String(bytes, offset, length) to do this is wrong, as it does not respect the correct character set and may
// return wrong results (depending on the platform's defaults)!
// BytesRef implements Comparable. The underlying byte arrays are sorted lexicographically, numerically treating
// elements as unsigned. This is identical to Unicode codepoint order.
type BytesRef struct {
	bs     []byte // The contents of the BytesRef. Should never be null.
	offset int    // offset of first valid byte.
	length int    // length of used bytes.
}

type bytesRefOption struct {
	bs     []byte
	offset int
	length int
}

type BytesRefOption func(o *bytesRefOption)

func NewBytesRef(options ...BytesRefOption) *BytesRef {
	opt := bytesRefOption{
		bs:     EMPTY_BYTES,
		offset: 0,
		length: 0,
	}
	for _, fn := range options {
		fn(&opt)
	}

	if opt.offset == 0 && opt.length == 0 {
		opt.offset = 0
		opt.length = len(opt.bs)
	}

	return &BytesRef{
		bs:     opt.bs,
		offset: opt.offset,
		length: opt.length,
	}
}

func (b *BytesRef) Bytes() []byte {
	return b.bs[b.offset : b.offset+b.length]
}

type BytesIterator interface {
	// Next Increments the iteration to the next BytesRef in the iterator. Returns the resulting BytesRef or
	// null if the end of the iterator is reached. The returned BytesRef may be re-used across calls to next.
	// After this method returns null, do not call it again: the results are undefined.
	// Returns: the next BytesRef in the iterator or null if the end of the iterator is reached.
	// Throws: 	IOException – If there is a low-level I/O error.
	Next() ([]byte, error)
}

func BytesToString(values []byte) string {
	sb := new(bytes.Buffer)

	sb.WriteByte('[')

	for i, value := range values {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(fmt.Sprintf("0x%x", value))
	}

	sb.WriteByte(']')
	return sb.String()
}

func StringToBytes(value string) ([]byte, error) {
	if len(value) < 2 {
		return nil, fmt.Errorf("string '%s'  was not created from BytesToString", value)
	}

	if !strings.HasPrefix(value, "[") || !strings.HasSuffix(value, "]") {
		return nil, fmt.Errorf("string '%s'  was not created from BytesToString", value)
	}

	parts := strings.Split(value[1:len(value)-1], " ")

	bs := make([]byte, 0, len(parts))
	for _, part := range parts {
		parseInt, err := strconv.ParseInt(strings.ReplaceAll(part, "0x", ""), 16, 16)
		if err != nil {
			return nil, err
		}
		bs = append(bs, byte(parseInt))
	}
	return bs, nil
}
