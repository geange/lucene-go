package fst

import (
	"context"
	"errors"
	"fmt"
)

// NodeHash Used to dedup states (lookup already-frozen states)
type NodeHash[T any] struct {
	table      map[int]int64
	fst        *FST[T]
	scratchArc *Arc[T]
	in         BytesReader
}

func NewNodeHash[T any](fst *FST[T], in BytesReader) *NodeHash[T] {
	return &NodeHash[T]{
		table:      make(map[int]int64),
		fst:        fst,
		scratchArc: &Arc[T]{},
		in:         in,
	}
}

const (
	PRIME = int64(32)
)

// 计算从当前节点起，后续的节点是否相同
func (n *NodeHash[T]) nodesEqual(ctx context.Context, node *UnCompiledNode[T], address int64) bool {
	if _, err := n.fst.ReadFirstRealTargetArc(ctx, address, n.in, n.scratchArc); err != nil {
		return false
	}

	// Fail fast for a node with fixed length arcs.
	if n.scratchArc.BytesPerArc() != 0 {
		if n.scratchArc.NodeFlags() == ArcsForBinarySearch {
			if node.NumArcs() != n.scratchArc.NumArcs() {
				return false
			}
		} else {
			if n.scratchArc.NodeFlags() != ArcsForDirectAddressing {
				panic("")
			}

			if node.Arcs[len(node.Arcs)-1].Label-node.Arcs[0].Label+1 != n.scratchArc.NumArcs() {
				return false
			} else if v, err := CountBits(n.scratchArc, n.in); err == nil && v != node.NumArcs() {
				return false
			}
		}
	}

	lastIdx := node.NumArcs() - 1

	for i, arc := range node.Arcs {
		if arc.Label != n.scratchArc.Label() ||
			!n.fst.outputs.Equal(arc.Output, n.scratchArc.output) ||
			arc.Target.(*CompiledNode).node != n.scratchArc.Target() ||
			!n.fst.outputs.Equal(arc.NextFinalOutput, n.scratchArc.NextFinalOutput()) ||
			arc.IsFinal != n.scratchArc.IsFinal() {

			return false
		}

		if n.scratchArc.IsLast() {
			if i == lastIdx {
				return true
			}
			return false
		}

		if _, err := n.fst.ReadNextRealArc(ctx, n.in, n.scratchArc); err != nil {
			return false
		}
	}

	return false
}

// hash code for an unfrozen node.
// This must be identical to the frozen case (below)!!
func (n *NodeHash[T]) hash(node *UnCompiledNode[T]) (int64, error) {
	h := int64(0)
	// TODO: maybe if number of arcs is high we can safely subsample?

	for _, arc := range node.Arcs {
		h = PRIME*h + int64(arc.Label)

		target, ok := arc.Target.(*CompiledNode)
		if !ok {
			return 0, errors.New("target is not CompiledNode")
		}

		nodeValue := target.node

		h = PRIME*h + (nodeValue ^ (nodeValue >> 32))
		h = PRIME*h + n.fst.outputs.Hash(arc.Output)
		h = PRIME*h + n.fst.outputs.Hash(arc.NextFinalOutput)
		if arc.IsFinal {
			h += 17
		}
	}
	return h, nil
}

func (n *NodeHash[T]) hashFrozenNode(ctx context.Context, node int64) (int64, error) {
	h := int64(0)
	if _, err := n.fst.ReadFirstRealTargetArc(ctx, node, n.in, n.scratchArc); err != nil {
		return 0, err
	}

	for {
		h = PRIME*h + int64(n.scratchArc.Label())
		h = PRIME*h + (n.scratchArc.Target() ^ (n.scratchArc.Target() >> 32))
		h = PRIME*h + n.fst.outputs.Hash(n.scratchArc.Output())
		h = PRIME*h + n.fst.outputs.Hash(n.scratchArc.NextFinalOutput())

		if n.scratchArc.IsFinal() {
			h += 17
		}

		if n.scratchArc.IsLast() {
			break
		}
		if _, err := n.fst.ReadNextRealArc(ctx, n.in, n.scratchArc); err != nil {
			return 0, err
		}
	}

	return h, nil
}

func (n *NodeHash[T]) Add(ctx context.Context, builder *Builder[T], nodeIn *UnCompiledNode[T]) (int64, error) {
	h, err := n.hash(nodeIn)
	if err != nil {
		return 0, err
	}

	pos := int(h)

	for {
		v, ok := n.table[pos]
		if !ok {
			// freeze & add
			node, err := n.fst.AddNode(ctx, builder, nodeIn)
			if err != nil {
				return 0, err
			}

			frozenNode, err := n.hashFrozenNode(ctx, node)
			if err != nil {
				return 0, err
			}

			if frozenNode != h {
				return 0, fmt.Errorf("frozenHash=%d vs h=%d", frozenNode, h)
			}

			n.table[pos] = node

			return node, nil
		}

		// 如果存在已有相同后缀的节点，找到map中的已有的节点，直接返回，保证后缀复用
		if n.nodesEqual(ctx, nodeIn, v) {
			return v, nil
		}
		pos++
	}
}

// called only by rehash
func (n *NodeHash[T]) addNew(ctx context.Context, address int64) error {
	v, err := n.hashFrozenNode(ctx, address)
	if err != nil {
		return err
	}

	//pos := v & n.mask
	pos := int(v)
	//c := int64(0)
	for {
		if _, ok := n.table[pos]; ok {
			n.table[pos] = address
			break
		}

		// quadratic probe
		//pos = (pos + (c + 1)) & n.mask
		//c++
		pos++
	}
	return nil
}
