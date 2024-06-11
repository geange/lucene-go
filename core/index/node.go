package index

import (
	"errors"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
)

type NodeApply interface {
	Apply(bufferedDeletes *BufferedUpdates, docIDUpto int) error
	IsDelete() bool
}

type Node struct {
	next  *Node
	item  any
	apply NodeApply
}

func NewNode(item any, apply NodeApply) *Node {
	return &Node{
		next:  nil,
		item:  item,
		apply: apply,
	}
}

func (n *Node) Apply(bufferedDeletes *BufferedUpdates, docIDUpto int) error {
	return n.apply.Apply(bufferedDeletes, docIDUpto)
}

var _ NodeApply = &TermNode{}

type TermNode struct {
	item index.Term
}

func NewTermNode(item index.Term) *TermNode {
	return &TermNode{item: item}
}

func (t *TermNode) Apply(bufferedDeletes *BufferedUpdates, docIDUpto int) error {
	bufferedDeletes.AddTerm(t.item, docIDUpto)
	return nil
}

func (t *TermNode) IsDelete() bool {
	return true
}

var _ NodeApply = &DocValuesUpdatesNode{}

type DocValuesUpdatesNode struct {
	updates []DocValuesUpdate
}

func NewDocValuesUpdatesNode(updates []DocValuesUpdate) *DocValuesUpdatesNode {
	return &DocValuesUpdatesNode{updates: updates}
}

func (d *DocValuesUpdatesNode) Apply(bufferedDeletes *BufferedUpdates, docIDUpto int) error {
	for _, update := range d.updates {
		switch update.GetType() {
		case document.DOC_VALUES_TYPE_NUMERIC:
			return bufferedDeletes.AddNumericUpdate(update.(*NumericDocValuesUpdate), docIDUpto)
		case document.DOC_VALUES_TYPE_BINARY:
			return bufferedDeletes.AddBinaryUpdate(update.(*BinaryDocValuesUpdate), docIDUpto)
		default:
			return errors.New("type not supported yet")
		}
	}
	return nil
}

func (d *DocValuesUpdatesNode) IsDelete() bool {
	return true
}
