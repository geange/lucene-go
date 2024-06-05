package query

import (
	"encoding/binary"
	"errors"
	"math"

	"github.com/geange/lucene-go/core/document"
)

func encodeFloat64(val float64, dst []byte, offset int) {
	value := math.Float64bits(val) ^ 0x8000000000000000
	binary.BigEndian.PutUint64(dst[offset:], value)
}

func verifyAndEncodeFloat64(mins, maxs []float64, dst []byte) error {
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

		encodeFloat64(mins[d], dst, i)
		encodeFloat64(maxs[d], dst, j)

		d++
		i += document.LONG_BYTES
		j += document.LONG_BYTES
	}

	return nil
}
