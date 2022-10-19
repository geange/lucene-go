package fst

import (
	"github.com/geange/lucene-go/core/util"
	"reflect"
)

// Node NOTE: not many instances of Node or CompiledNode are
// in memory while the FST is being built; it's only the
// current "frontier":
type Node interface {
	IsCompiled() bool
}

type CompiledNode struct {
	node int64
}

func (*CompiledNode) IsCompiled() bool {
	return true
}

type UnCompiledNode struct {
	owner   *Builder
	numArcs int
	arcs    []*Arc

	// TODO: instead of recording isFinal/output on the
	// node, maybe we should use -1 arc to mean "end" (like
	// we do when reading the FST).  Would simplify much
	// code here...
	output     any
	isFinal    bool
	inputCount int

	// This node's depth, starting from the automaton root.
	depth int
}

func NewUnCompiledNode(owner *Builder, depth int) *UnCompiledNode {
	return &UnCompiledNode{
		owner:  owner,
		arcs:   []*Arc{{}},
		output: owner.NO_OUTPUT,
		depth:  depth,
	}
}

func (*UnCompiledNode) IsCompiled() bool {
	return false
}

func (u *UnCompiledNode) Clear() {
	u.numArcs = 0
	u.isFinal = false
	u.output = u.owner.NO_OUTPUT
	u.inputCount = 0
}

func (u *UnCompiledNode) GetLastOutput(labelToMatch int) any {
	return u.arcs[u.numArcs-1].output
}

func (u *UnCompiledNode) AddArc(label int, target Node) {
	if u.numArcs == len(u.arcs) {
		newArcs := util.Grow(u.arcs, u.numArcs+1)
		for arcIdx := u.numArcs; arcIdx < len(newArcs); arcIdx++ {
			newArcs[arcIdx] = &Arc{}
		}
		u.arcs = newArcs
	}
	arc := u.arcs[u.numArcs]
	u.numArcs++

	arc.label = label
	arc.target = target
	arc.output, arc.nextFinalOutput = u.owner.NO_OUTPUT, u.owner.NO_OUTPUT
	arc.isFinal = false
}

func (u *UnCompiledNode) ReplaceLast(labelToMatch int, target Node, nextFinalOutput any, isFinal bool) {
	// assert numArcs > 0;
	arc := u.arcs[u.numArcs-1]
	// assert arc.label == labelToMatch: "arc.label=" + arc.label + " vs " + labelToMatch;
	arc.target = target
	//assert target.node != -2;
	arc.nextFinalOutput = nextFinalOutput
	arc.isFinal = isFinal
}

func (u *UnCompiledNode) DeleteLast(label int, target Node) error {
	err := assert(u.numArcs > 0)
	if err != nil {
		return err
	}
	err = assert(label == u.arcs[u.numArcs-1].label)
	if err != nil {
		return err
	}
	err = assert(reflect.DeepEqual(target, u.arcs[u.numArcs-1].target))
	if err != nil {
		return err
	}

	u.numArcs--
	return nil
}

func (u *UnCompiledNode) SetLastOutput(labelToMatch int, newOutput any) error {
	//assert owner.validOutput(newOutput);
	//assert(u.numArcs > 0)

	arc := u.arcs[u.numArcs-1]
	//assert(arc.label == labelToMatch)
	arc.output = newOutput
	return nil
}

// PrependOutput pushes an output prefix forward onto all arcs
func (u *UnCompiledNode) PrependOutput(outputPrefix any) (err error) {
	//  assert owner.validOutput(outputPrefix);
	for arcIdx := 0; arcIdx < u.numArcs; arcIdx++ {
		u.arcs[arcIdx].output, err = u.owner.fst.outputs.Add(outputPrefix, u.arcs[arcIdx].output)
		if err != nil {
			return err
		}
		//assert owner.validOutput(u.arcs[arcIdx].output);
	}

	if u.isFinal {
		u.output, err = u.owner.fst.outputs.Add(outputPrefix, u.output)
		if err != nil {
			return err
		}
		//assert owner.validOutput(output);
	}

	return nil
}
