package fst

import (
	"fmt"
	"math"
	"reflect"

	"github.com/geange/lucene-go/core/util/packed"
)

// NodeHash Used to dedup states (lookup already-frozen states)
type NodeHash struct {
	table      *packed.PagedGrowableWriter
	count      int64
	mask       int64
	fst        *FST
	scratchArc *FSTArc
	in         BytesReader
}

const (
	PRIME = int64(32)
)

func (n *NodeHash) nodesEqual(node *UnCompiledNode, address int64) bool {
	_, err := n.fst.ReadFirstRealTargetArc(address, n.scratchArc, n.in)
	if err != nil {
		return false
	}

	// Fail fast for a node with fixed length arcs.
	if n.scratchArc.BytesPerArc() != 0 {
		if n.scratchArc.NodeFlags() == ARCS_FOR_BINARY_SEARCH {
			if node.NumArcs != n.scratchArc.NumArcs() {
				return false
			}
		} else {
			{
				if n.scratchArc.NodeFlags() != ARCS_FOR_DIRECT_ADDRESSING {
					panic("")
				}
			}

			if int64(node.Arcs[len(node.Arcs)-1].Label-node.Arcs[0].Label+1) != n.scratchArc.NumArcs() {
				return false
			} else if v, err := BitTable.countBits(n.scratchArc, n.in); err == nil && v != node.NumArcs {
				return false
			}
		}
	}

	for i := range node.Arcs {
		arc := node.Arcs[i]
		if arc.Label != n.scratchArc.Label() ||
			!(reflect.DeepEqual(arc.Output, n.scratchArc.output)) ||
			arc.Target.(*CompiledNode).node != n.scratchArc.NextFinalOutput() ||
			arc.IsFinal != n.scratchArc.IsFinal() {
			return false
		}

		if n.scratchArc.IsLast() {
			if i == int(node.NumArcs-1) {
				return true
			}
			return false
		}

		_, err := n.fst.ReadNextRealArc(n.scratchArc, n.in)
		if err != nil {
			return false
		}
	}

	return false
}

// hashNode code for an unfrozen node.  This must be identical
// to the frozen case (below)!!
func (n *NodeHash) hashUnfrozenNode(node *UnCompiledNode) int64 {
	h := int64(0)
	// TODO: maybe if number of arcs is high we can safely subsample?

	for i := range node.Arcs {
		arc := node.Arcs[i]
		h = PRIME*h + int64(arc.Label)

		n := arc.Target.(*CompiledNode).node

		h = PRIME*h + (n ^ (n >> 32))
		h = PRIME*h + hashObj(arc.Output)
		h = PRIME*h + hashObj(arc.NextFinalOutput)
		if arc.IsFinal {
			h += 17
		}
	}
	return h & math.MaxInt64
}

func (n *NodeHash) hashFrozenNode(node int64) (int64, error) {
	h := int64(0)
	var err error
	_, err = n.fst.ReadFirstRealTargetArc(node, n.scratchArc, n.in)
	if err != nil {
		return 0, err
	}

	for {
		h = PRIME*h + int64(n.scratchArc.Label())
		h = PRIME*h + (n.scratchArc.Target() ^ (n.scratchArc.Target() >> 32))
		h = PRIME*h + hashObj(n.scratchArc.Output())
		h = PRIME*h + hashObj(n.scratchArc.NextFinalOutput())

		if n.scratchArc.IsFinal() {
			h += 17
		}

		if n.scratchArc.IsLast() {
			break
		}
		_, err := n.fst.ReadNextRealArc(n.scratchArc, n.in)
		if err != nil {
			return 0, err
		}
	}

	return h & math.MaxInt64, nil
}

func (n *NodeHash) Add(builder *Builder, nodeIn *UnCompiledNode) (int64, error) {
	h := n.hashUnfrozenNode(nodeIn)
	pos := h & n.mask
	c := int64(0)

	for {
		v := n.table.Get(int(pos))
		if v == 0 {
			// freeze & add
			node, err := n.fst.AddNode(builder, nodeIn)
			if err != nil {
				return 0, err
			}

			{
				frozenNode, err := n.hashFrozenNode(node)
				if err != nil {
					return 0, err
				}
				if frozenNode != h {
					return 0, fmt.Errorf("frozenHash=%d vs h=%d", frozenNode, h)
				}
			}

			n.count++
			n.table.Set(int(pos), uint64(node))
			// Rehash at 2/3 occupancy:
			if n.count > int64(2*n.table.Size()/3) {
				err := n.rehash()
				if err != nil {
					return 0, err
				}
			}
			return node, nil
		}
		if n.nodesEqual(nodeIn, int64(v)) {
			return int64(v), nil
		}

		pos = (pos + (c + 1)) & n.mask
		c++
	}
}

// called only by rehash
func (n *NodeHash) addNew(address int64) error {
	v, err := n.hashFrozenNode(address)
	if err != nil {
		return err
	}
	pos := v & n.mask
	c := int64(0)
	for {
		if n.table.Get(int(pos)) == 0 {
			n.table.Set(int(pos), uint64(address))
			break
		}

		// quadratic probe
		pos = (pos + (c + 1)) & n.mask
		c++
	}
	return nil
}

func (n *NodeHash) rehash() error {
	oldTable := n.table
	var err error
	n.table, err = packed.NewPagedGrowableWriter(
		2*oldTable.Size(), 1<<30, packed.PackedIntsBitsRequired(uint64(n.count)), packed.COMPACT)
	if err != nil {
		return err
	}
	n.mask = int64(n.table.Size() - 1)
	for i := 0; i < oldTable.Size(); i++ {
		address := oldTable.Get(i)
		if address != 0 {
			err := n.addNew(int64(address))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func hashObj(obj interface{}) int64 {
	// TODO: != NO_OUTPUT
	if obj != nil {
		switch obj.(type) {
		case []byte:
			h := int64(0)
			for _, b := range obj.([]byte) {
				h = PRIME*h + int64(b)
			}
			return h
		}

	}
	return -1
}
