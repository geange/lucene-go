package query

import (
	"encoding/binary"
	"errors"

	"github.com/geange/lucene-go/core/document"
)

func encodeInt32(val int32, dst []byte, offset int) {
	value := uint32(val) ^ 0x80000000
	binary.BigEndian.PutUint32(dst[offset:], value)
}

func verifyAndEncodeInt32(mins, maxs []int32, dst []byte) error {
	for d, i, j := 0, 0, len(mins)*document.INTEGER_BYTES; d < len(mins); {

		if IsNaN(mins[d]) {
			return errors.New("invalid min value")
		}

		if IsNaN(maxs[d]) {
			return errors.New("invalid max value")
		}

		if mins[d] > maxs[d] {
			return errors.New("min value is greater than max value")
		}

		encodeInt32(mins[d], dst, i)
		encodeInt32(maxs[d], dst, j)

		d++
		i += document.INTEGER_BYTES
		j += document.INTEGER_BYTES
	}

	return nil
}
