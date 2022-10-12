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

/**

  public long add(Builder<T> builder, Builder.UnCompiledNode<T> nodeIn) throws IOException {
    final long h = hash(nodeIn);
    long pos = h & mask;
    int c = 0;
    while(true) {
      final long v = table.get(pos);
      if (v == 0) {
        final long node = fst.addNode(builder, nodeIn);
        count++;
        table.set(pos, node);
        // Rehash at 2/3 occupancy:
        if (count > 2*table.size()/3) {
          rehash();
        }
        return node;
      } else if (nodesEqual(nodeIn, v)) {
        // same node is already here
        return v;
      }

      // quadratic probe
      pos = (pos + (++c)) & mask;
    }
  }


*/

func (n *NodeHash[T]) Add(builder *Builder[T], nodeIn *UnCompiledNode[T]) (int, error) {
	panic("")
}

// hash code for an unfrozen node.  This must be identical
// to the frozen case (below)!!
func (n *NodeHash[T]) hash(node *UnCompiledNode[T]) int64 {
	panic("")
}

/**

  private long hash(long node) throws IOException {
    final int PRIME = 31;
    //System.out.println("hash frozen node=" + node);
    long h = 0;
    fst.readFirstRealTargetArc(node, scratchArc, in);
    while(true) {
      // System.out.println("  label=" + scratchArc.label + " target=" + scratchArc.target + " h=" + h + " output=" + fst.outputs.outputToString(scratchArc.output) + " next?=" + scratchArc.flag(4) + " final?=" + scratchArc.isFinal() + " pos=" + in.getPosition());
      h = PRIME * h + scratchArc.label();
      h = PRIME * h + (int) (scratchArc.target() ^(scratchArc.target() >>32));
      h = PRIME * h + scratchArc.output().hashCode();
      h = PRIME * h + scratchArc.nextFinalOutput().hashCode();
      if (scratchArc.isFinal()) {
        h += 17;
      }
      if (scratchArc.isLast()) {
        break;
      }
      fst.readNextRealArc(scratchArc, in);
    }
    //System.out.println("  ret " + (h&Integer.MAX_VALUE));
    return h & Long.MAX_VALUE;
  }

*/

func (n *NodeHash[T]) hashFrozen(node int) (int, error) {
	PRIME := 31
	h := 0
	_, err := n.fst.ReadFirstRealTargetArc(node, n.scratchArc, n.in)
	if err != nil {
		return 0, err
	}

	for {
		h = PRIME*h + n.scratchArc.Label()
		h = PRIME*h + (int)(n.scratchArc.Target()^(n.scratchArc.Target()>>32))
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

/**

  private void addNew(long address) throws IOException {
    long pos = hash(address) & mask;
    int c = 0;
    while(true) {
      if (table.get(pos) == 0) {
        table.set(pos, address);
        break;
      }

      // quadratic probe
      pos = (pos + (++c)) & mask;
    }
  }


*/

func (n *NodeHash[T]) addNew(address int) error {
	frozen, err := n.hashFrozen(address)
	c := 0
	if err != nil {
		return err
	}
	pos := frozen & n.mask
	for {
		if n.table.Get(pos) == 0 {
			n.table.Set(pos, int64(address))
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
	n.table = packed.NewPagedGrowableWriter(2*oldTable.Size(), 1<<30,
		packed.PackedIntsBitsRequired(int64(n.count)), packed.COMPACT)
	n.mask = n.table.Size() - 1
	for idx := 0; idx < oldTable.Size(); idx++ {
		address := oldTable.Get(idx)
		if address != 0 {
			return n.addNew(int(address))
		}
	}
	return nil
}
