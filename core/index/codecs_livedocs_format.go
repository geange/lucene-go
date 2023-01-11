package index

import (
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
)

type LiveDocsFormat interface {

	// ReadLiveDocs Read live docs bits.
	ReadLiveDocs(dir store.Directory, info SegmentCommitInfo, context *store.IOContext) (util.Bits, error)

	// WriteLiveDocs Persist live docs bits. Use SegmentCommitInfo.getNextDelGen to determine
	// the generation of the deletes file you should write to.
	WriteLiveDocs(bits util.Bits, dir store.Directory,
		info *SegmentCommitInfo, newDelCount int, context *store.IOContext) error

	// Files Records all files in use by this SegmentCommitInfo into the files argument.
	Files(info *SegmentCommitInfo, files map[string]struct{}) (map[string]struct{}, error)
}
