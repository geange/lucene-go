package simpletext

import (
	"bytes"
	"github.com/geange/lucene-go/core/util/bkd"
)

const (
	CODEC_NAME                    = "BKD"
	VERSION_START                 = 0
	VERSION_COMPRESSED_DOC_IDS    = 1
	VERSION_COMPRESSED_VALUES     = 2
	VERSION_IMPLICIT_SPLIT_DIM_1D = 3
	VERSION_CURRENT               = VERSION_IMPLICIT_SPLIT_DIM_1D
	DEFAULT_MAX_MB_SORT_IN_HEAP   = 16.0
)

type SimpleTextBKDWriter struct {
	// How many dimensions we are storing at the leaf (data) nodes
	config *bkd.BKDConfig

	scratch *bytes.Buffer
}
