package blocktree

import "github.com/geange/lucene-go/core/store"

const (
	OUTPUT_FLAGS_NUM_BITS = 2
	OUTPUT_FLAGS_MASK     = 0x3
	OUTPUT_FLAG_IS_FLOOR  = 0x1
	OUTPUT_FLAG_HAS_TERMS = 0x2

	TERMS_EXTENSION             = "tim"
	TERMS_CODEC_NAME            = "BlockTreeTermsDict"
	VERSION_START               = 3
	VERSION_META_LONGS_REMOVED  = 4
	VERSION_COMPRESSED_SUFFIXES = 5
	VERSION_META_FILE           = 6
	VERSION_CURRENT             = VERSION_META_FILE
	TERMS_INDEX_EXTENSION       = "tip"
	TERMS_INDEX_CODEC_NAME      = "BlockTreeTermsIndex"
	TERMS_META_EXTENSION        = "tmd"
	TERMS_META_CODEC_NAME       = "BlockTreeTermsMeta"
)

type TermsReader struct {
	termsIn store.IndexInput // Open input to the main terms dict file (_X.tib)
	indexIn store.IndexInput // Open input to the terms index file (_X.tip)

	postingsReader
}
