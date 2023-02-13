package index

import "github.com/geange/lucene-go/core/store"

type TermVectorsConsumer struct {
	*TermsHash

	directory store.Directory
	info      *SegmentInfo
	codec     Codec
	writer    TermVectorsWriter
}
