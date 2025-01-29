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

func verifyAndEncodeInt64(minNums, maxNums []int64, dst []byte) error {
	for d, i, j := 0, 0, len(minNums)*document.LONG_BYTES; d < len(minNums); {

		if IsNaN(minNums[d]) {
			return errors.New("invalid min value")
		}

		if IsNaN(maxNums[d]) {
			return errors.New("invalid max value")
		}

		if minNums[d] > maxNums[d] {
			return errors.New("min value is greater than max value")
		}

		encodeInt64(minNums[d], dst, i)
		encodeInt64(maxNums[d], dst, j)

		d++
		i += document.LONG_BYTES
		j += document.LONG_BYTES
	}

	return nil
}
