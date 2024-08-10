package index

import (
	"errors"

	"github.com/geange/lucene-go/core/interface/index"
)

func TestLiveDocs(reader index.CodecReader) error {
	numDocs := reader.NumDocs()
	if reader.HasDeletions() {
		liveDocs := reader.GetLiveDocs()
		if liveDocs == nil {
			return errors.New("segment should have deletions, but liveDocs is null")
		}

		numLive := 0
		size := int(liveDocs.Len())
		for i := 0; i < size; i++ {
			if liveDocs.Test(uint(i)) {
				numLive++
			}
		}
		if numLive != numDocs {
			return errors.New("liveDocs count mismatch")
		}
		return nil
	}

	liveDocs := reader.GetLiveDocs()
	if liveDocs != nil {
		// it's ok for it to be non-null here, as long as none are set right?
		size := int(liveDocs.Len())
		for i := 0; i < size; i++ {
			if !liveDocs.Test(uint(i)) {
				return errors.New("liveDocs mismatch")
			}
		}
	}
	return nil
}
