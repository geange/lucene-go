package bytesref

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
)

var (
	EMPTY_BYTES []byte
)

type BytesIterator interface {
	// Next Increments the iteration to the next BytesRef in the iterator. Returns the resulting BytesRef or
	// null if the end of the iterator is reached. The returned BytesRef may be re-used across calls to next.
	// After this method returns null, do not call it again: the results are undefined.
	// Returns: the next BytesRef in the iterator or null if the end of the iterator is reached.
	// Throws: 	IOException – If there is a low-level I/O error.
	Next(ctx context.Context) ([]byte, error)
}

func BytesToString(values []byte) string {
	buf := new(bytes.Buffer)

	buf.WriteByte('[')

	for i, value := range values {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString("0x")
		buf.WriteString(fmt.Sprintf("%x", value))
	}

	buf.WriteByte(']')
	return buf.String()
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
