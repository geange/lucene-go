package fst

// Node NOTE: not many instances of Node or CompiledNode are in
// memory while the FST is being built; it's only the current "frontier":
type Node interface {
	IsCompiled() bool
}

var _ Node = &CompiledNode{}

type CompiledNode struct {
	node int64
}

func NewCompiledNode() *CompiledNode {
	return &CompiledNode{}
}

func (*CompiledNode) IsCompiled() bool {
	return true
}

var _ Node = &UnCompiledNode[any]{}

// UnCompiledNode Expert: holds a pending (seen but not yet serialized) Node.
type UnCompiledNode[T any] struct {
	owner *Builder[T]

	numArcs int
	arcs    []*builderArc[T]

	// TODO: instead of recording isFinal/output on the
	// node, maybe we should use -1 arc to mean "end" (like
	// we do when reading the FST).  Would simplify much
	// code here...
	output     *Box[T]
	isFinal    bool
	inputCount int

	// This node's depth, starting from the automaton root.
	depth int
}

func NewUnCompiledNode[T any](owner *Builder[T], depth int) *UnCompiledNode[T] {
	this := &UnCompiledNode[T]{
		owner:  owner,
		arcs:   make([]*builderArc[T], 1),
		output: owner.noOutput,
		depth:  depth,
	}
	this.arcs[0] = newBuilderArc[T]()
	return this
}

func (u *UnCompiledNode[T]) IsCompiled() bool {
	return false
}

func (u *UnCompiledNode[T]) Clear() {
	u.numArcs = 0
	u.isFinal = false
	u.output = u.owner.noOutput
	u.inputCount = 0
}

func (u *UnCompiledNode[T]) GetLastOutput(labelToMatch int) *Box[T] {
	return u.arcs[u.numArcs-1].output
}

/**

  public void addArc(int label, Node target) {
    assert label >= 0;
    assert numArcs == 0 || label > arcs[numArcs-1].label: "arc[numArcs-1].label=" + arcs[numArcs-1].label + " new label=" + label + " numArcs=" + numArcs;
    if (numArcs == arcs.length) {
      final Arc<T>[] newArcs = ArrayUtil.grow(arcs, numArcs+1);
      for(int arcIdx=numArcs;arcIdx<newArcs.length;arcIdx++) {
        newArcs[arcIdx] = new Arc<>();
      }
      arcs = newArcs;
    }
    final Arc<T> arc = arcs[numArcs++];
    arc.label = label;
    arc.target = target;
    arc.output = arc.nextFinalOutput = owner.NO_OUTPUT;
    arc.isFinal = false;
  }

*/

func (u *UnCompiledNode[T]) AddArc(label int, target Node) {
	if u.numArcs == len(u.arcs) {
		u.arcs = append(u.arcs, newBuilderArc[T]())
	}
	arc := u.arcs[u.numArcs]
	u.numArcs++
	arc.label = label
	arc.target = target
	arc.output, arc.nextFinalOutput = u.owner.noOutput, u.owner.noOutput
	arc.isFinal = false
}

/**

  public void replaceLast(int labelToMatch, Node target, T nextFinalOutput, boolean isFinal) {
    assert numArcs > 0;
    final Arc<T> arc = arcs[numArcs-1];
    assert arc.label == labelToMatch: "arc.label=" + arc.label + " vs " + labelToMatch;
    arc.target = target;
    //assert target.node != -2;
    arc.nextFinalOutput = nextFinalOutput;
    arc.isFinal = isFinal;
  }

*/

func (u *UnCompiledNode[T]) ReplaceLast(labelToMatch int, target Node, nextFinalOutput any, isFinal bool) {

}

func (u *UnCompiledNode[T]) DeleteLast(label int, target Node) {

}

func (u *UnCompiledNode[T]) SetLastOutput(labelToMatch int, newOutput any) {

}

/**
  public void prependOutput(T outputPrefix) {
    assert owner.validOutput(outputPrefix);

    for(int arcIdx=0;arcIdx<numArcs;arcIdx++) {
      arcs[arcIdx].output = owner.fst.outputs.add(outputPrefix, arcs[arcIdx].output);
      assert owner.validOutput(arcs[arcIdx].output);
    }

    if (isFinal) {
      output = owner.fst.outputs.add(outputPrefix, output);
      assert owner.validOutput(output);
    }
  }
*/

// PrependOutput pushes an output prefix forward onto all arcs
func (u *UnCompiledNode[T]) PrependOutput(outputPrefix *Box[T]) {
	for _, arc := range u.arcs {
		arc.output = u.owner.fst.outputs.Add(outputPrefix, arc.output)
	}

	if u.isFinal {
		u.output = u.owner.fst.outputs.Add(outputPrefix, u.output)
	}
}
