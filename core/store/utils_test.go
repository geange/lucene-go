package store

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSegmentFileName(t *testing.T) {
	items := []struct {
		segmentName, segmentSuffix, ext, result string
	}{
		{
			segmentName:   "a",
			segmentSuffix: "",
			ext:           "ext",
			result:        "a.ext",
		},
		{
			segmentName:   "a",
			segmentSuffix: "b",
			ext:           "ext",
			result:        "a_b.ext",
		},
		{
			segmentName:   "aaaaaaaa",
			segmentSuffix: "b",
			ext:           "ext",
			result:        "aaaaaaaa_b.ext",
		},
	}

	for _, v := range items {
		name := SegmentFileName(v.segmentName, v.segmentSuffix, v.ext)
		assert.Equal(t, v.result, name)
	}
}
