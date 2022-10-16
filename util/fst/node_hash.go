package fst

import (
	"github.com/geange/lucene-go/core/util/packed"
	"math"
)

type NodeHash[T any] struct {
	table *packed.PagedGrowableWriter

	count int
	mask  int

	fst *FST[T]

	scratchArc *Arc[T]
	in         BytesReader
}

func NewNodeHash[T any](fst *FST[T], in BytesReader) (*NodeHash[T], error) {
	table, err := packed.NewPagedGrowableWriter(16, 1<<27, 8, packed.COMPACT)
	if err != nil {
		return nil, err
	}
	return &NodeHash[T]{
		table:      table,
		mask:       15,
		fst:        fst,
		in:         in,
		scratchArc: &Arc[T]{},
	}, nil
}

func (n *NodeHash[T]) Add(builder *Builder[T], nodeIn *UnCompiledNode[T]) (int, error) {
	h := n.hash(nodeIn)
	pos := h & int64(n.mask)
	c := 0

	for {
		v := n.table.Get(int(pos))
		if v == 0 {
			// freeze & add
			node, err := n.fst.addNode(builder, nodeIn)
			if err != nil {
				return 0, err
			}
			n.count++
			n.table.Set(int(pos), uint64(node))

			// Rehash at 2/3 occupancy:
			if n.count > 2*n.table.Size()/3 {
				err := n.rehash()
				if err != nil {
					return 0, err
				}
			}
			return node, nil
		} else if ok, err := n.nodesEqual(nodeIn, v); err != nil {
			return 0, err
		} else if ok {
			return int(v), nil
		}

		// quadratic probe
		pos = (pos + (int64(c) + 1)) & int64(n.mask)
	}
}

func (n *NodeHash[T]) nodesEqual(node *UnCompiledNode[T], address uint64) (bool, error) {
	_, err := n.fst.ReadFirstRealTargetArc(int(address), n.scratchArc, n.in)
	if err != nil {
		return false, err
	}

	// Fail fast for a node with fixed length arcs.
	if n.scratchArc.BytesPerArc() != 0 {
		if n.scratchArc.NodeFlags() == ARCS_FOR_BINARY_SEARCH {
			if node.numArcs != n.scratchArc.NumArcs() {
				return false, nil
			}
		} else {
			if (node.arcs[node.numArcs-1].label-node.arcs[0].label+1) != n.scratchArc.NumArcs() ||
				node.numArcs != CountBits(n.scratchArc, n.in) {
				return false, nil
			}
		}
	}

	for arcUpto := 0; arcUpto < node.numArcs; arcUpto++ {
		arc := node.arcs[arcUpto]

		if int64(arc.label) != n.scratchArc.Label() ||
			arc.isFinal != n.scratchArc.IsFinal() {
			return false, nil
		}

		if n.scratchArc.IsLast() {
			if arcUpto == node.numArcs-1 {
				return true, nil
			} else {
				return false, nil
			}
		}

		_, err := n.fst.ReadNextRealArc(n.scratchArc, n.in)
		if err != nil {
			return false, err
		}
	}

	return false, nil
}

// hash code for an unfrozen node.  This must be identical
// to the frozen case (below)!!
func (n *NodeHash[T]) hash(node *UnCompiledNode[T]) int64 {
	panic("")
}

func (n *NodeHash[T]) hashFrozen(node int) (int, error) {
	PRIME := 31
	h := 0
	_, err := n.fst.ReadFirstRealTargetArc(node, n.scratchArc, n.in)
	if err != nil {
		return 0, err
	}

	for {
		h = PRIME*h + int(n.scratchArc.Label())
		h = PRIME*h + (n.scratchArc.Target() ^ (n.scratchArc.Target() >> 32))
		h = PRIME*h + n.scratchArc.Output().hashCode()
		h = PRIME*h + n.scratchArc.NextFinalOutput().hashCode()
		if n.scratchArc.IsFinal() {
			h += 17
		}
		if n.scratchArc.IsLast() {
			break
		}
		n.fst.ReadNextRealArc(n.scratchArc, n.in)
	}

	return h & math.MaxInt64, nil
}

func (n *NodeHash[T]) addNew(address int) error {
	frozen, err := n.hashFrozen(address)
	c := 0
	if err != nil {
		return err
	}
	pos := frozen & n.mask
	for {
		if n.table.Get(pos) == 0 {
			n.table.Set(pos, uint64(address))
			break
		}

		// quadratic probe
		c++
		pos = (pos + c) & n.mask
	}
	return nil
}

func (n *NodeHash[T]) rehash() error {
	oldTable := n.table
	var err error
	n.table, err = packed.NewPagedGrowableWriter(2*oldTable.Size(), 1<<30,
		packed.PackedIntsBitsRequired(uint64(n.count)), packed.COMPACT)
	if err != nil {
		return err
	}
	n.mask = n.table.Size() - 1
	for idx := 0; idx < oldTable.Size(); idx++ {
		address := oldTable.Get(idx)
		if address != 0 {
			return n.addNew(int(address))
		}
	}
	return nil
}
