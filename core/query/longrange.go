package query

import (
	"encoding/binary"
	"errors"
	"github.com/geange/lucene-go/core/document"
)

func encodeInt64(val int64, dst []byte, offset int) {
	value := uint64(val) ^ 0x8000000000000000
	binary.BigEndian.PutUint64(dst[offset:], value)
}

func verifyAndEncodeInt64(mins, maxs []int64, dst []byte) error {
	for d, i, j := 0, 0, len(mins)*document.LONG_BYTES; d < len(mins); {

		if IsNaN(mins[d]) {
			return errors.New("invalid min value")
		}

		if IsNaN(maxs[d]) {
			return errors.New("invalid max value")
		}

		if mins[d] > maxs[d] {
			return errors.New("min value is greater than max value")
		}

		encodeInt64(mins[d], dst, i)
		encodeInt64(maxs[d], dst, j)

		d++
		i += document.LONG_BYTES
		j += document.LONG_BYTES
	}

	return nil
}
