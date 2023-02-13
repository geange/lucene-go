package index

type DocumentsWriterDeleteQueue struct {
}

type DeleteSlice struct {
}

type Node interface {
	Apply(bufferedDeletes *BufferedUpdates, docIDUpto int)
}

type TermNode struct {
	next *TermNode
	item *Term
}

func (t *TermNode) Apply(bufferedDeletes *BufferedUpdates, docIDUpto int) {
	bufferedDeletes.AddTerm(t.item, docIDUpto)
}
