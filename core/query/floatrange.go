package query

import (
	"encoding/binary"
	"errors"
	"math"

	"github.com/geange/lucene-go/core/document"
)

func encodeFloat32(val float32, dst []byte, offset int) {
	value := math.Float32bits(val) ^ 0x80000000
	binary.BigEndian.PutUint32(dst[offset:], value)
}

func verifyAndEncodeFloat32(minNums, maxNums []float32, dst []byte) error {
	for d, i, j := 0, 0, len(minNums)*document.INTEGER_BYTES; d < len(minNums); {

		if IsNaN(minNums[d]) {
			return errors.New("invalid min value")
		}

		if IsNaN(maxNums[d]) {
			return errors.New("invalid max value")
		}

		if minNums[d] > maxNums[d] {
			return errors.New("min value is greater than max value")
		}

		encodeFloat32(minNums[d], dst, i)
		encodeFloat32(maxNums[d], dst, j)

		d++
		i += document.INTEGER_BYTES
		j += document.INTEGER_BYTES
	}

	return nil
}
