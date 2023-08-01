package bkd

import "github.com/geange/lucene-go/core/store"

type docIdsWriter struct {
}

func (*docIdsWriter) writeDocIds(docIds []int, start, count int, out store.DataOutput) error {
	panic("")
}

// Read count integers into docIDs.
func (*docIdsWriter) readInts(in store.IndexInput, count int, docIDs []int) error {
	panic("")
}
