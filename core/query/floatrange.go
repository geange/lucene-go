package query

import (
	"encoding/binary"
	"errors"
	"github.com/geange/lucene-go/core/document"
	"math"
)

func encodeFloat32(val float32, dst []byte, offset int) {
	value := math.Float32bits(val) ^ 0x80000000
	binary.BigEndian.PutUint32(dst[offset:], value)
}

func verifyAndEncodeFloat32(mins, maxs []float32, dst []byte) error {
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

		encodeFloat32(mins[d], dst, i)
		encodeFloat32(maxs[d], dst, j)

		d++
		i += document.INTEGER_BYTES
		j += document.INTEGER_BYTES
	}

	return nil
}
