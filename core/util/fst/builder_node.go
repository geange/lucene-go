package fst

import (
	"errors"
	"fmt"
)

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

type UnCompiledNode[T any] struct {
	Owner *Builder[T]
	//NumArcs int64

	// TODO: instead of recording isFinal/output on the
	// node, maybe we should use -1 arc to mean "end" (like
	// we do when reading the FST).  Would simplify much
	// code here...
	Arcs       []*BuilderArc[T]
	Output     T
	IsFinal    bool
	InputCount int64

	// This node's depth, starting from the automaton root.
	Depth int
}

func (u *UnCompiledNode[T]) NumArcs() int64 {
	return int64(len(u.Arcs))
}

func NewUnCompiledNode[T any](owner *Builder[T], depth int) *UnCompiledNode[T] {
	return &UnCompiledNode[T]{
		Owner:  owner,
		Arcs:   make([]*BuilderArc[T], 0),
		Output: owner.noOutput,
		Depth:  depth,
	}
}

func (u *UnCompiledNode[T]) IsCompiled() bool {
	return false
}

func (u *UnCompiledNode[T]) Clear() {
	u.Arcs = u.Arcs[:0]
	u.IsFinal = false
	u.Output = u.Owner.noOutput
	u.InputCount = 0

	// We don't clear the depth here because it never changes
	// for nodes on the frontier (even when reused).
}

func (u *UnCompiledNode[T]) GetLastOutput(labelToMatch int) T {
	return u.Arcs[len(u.Arcs)-1].Output
}

func (u *UnCompiledNode[T]) AddArc(label int, target Node) {
	u.Arcs = append(u.Arcs, &BuilderArc[T]{
		Label:           label,
		Target:          target,
		Output:          u.Owner.noOutput,
		NextFinalOutput: u.Owner.noOutput,
		IsFinal:         false,
	})
}

func (u *UnCompiledNode[T]) DeleteLast(label int, target Node) error {
	if len(u.Arcs) <= 0 {
		return errors.New("arcs size is 0")
	}

	if label != u.Arcs[len(u.Arcs)-1].Label {
		return errors.New("label not match")
	}

	if target != u.Arcs[len(u.Arcs)-1].Target {
		return errors.New("target not match")
	}
	u.Arcs = u.Arcs[:len(u.Arcs)-1]
	return nil
}

func (u *UnCompiledNode[T]) SetLastOutput(label int, newOutput T) error {
	if len(u.Arcs) <= 0 {
		return errors.New("arcs size is 0")
	}
	arc := u.Arcs[len(u.Arcs)-1]
	if arc.Label != label {
		return errors.New("label not match")
	}
	arc.Output = newOutput
	return nil
}

func (u *UnCompiledNode[T]) ReplaceLast(labelToMatch int, target Node, nextFinalOutput T, isFinal bool) error {
	if len(u.Arcs) <= 0 {
		return fmt.Errorf("arcs size is 0")
	}

	arc := u.Arcs[len(u.Arcs)-1]
	if arc.Label != labelToMatch {
		return fmt.Errorf("arc.label=%d vs %d", arc.Label, labelToMatch)
	}
	arc.Target = target
	arc.NextFinalOutput = nextFinalOutput
	arc.IsFinal = isFinal
	return nil
}

// PrependOutput pushes an output prefix forward onto all arcs
func (u *UnCompiledNode[T]) PrependOutput(outputPrefix T) error {
	var err error
	for i := range u.Arcs {
		u.Arcs[i].Output, err = u.Owner.fst.outputs.Add(outputPrefix, u.Arcs[i].Output)
		if err != nil {
			return err
		}
		// TODO owner.validOutput(output)
	}

	if u.IsFinal {
		u.Output, err = u.Owner.fst.outputs.Add(outputPrefix, u.Output)
		if err != nil {
			return err
		}

		// TODO owner.validOutput(output)
	}

	return nil
}

// BuilderArc Expert: holds a pending (seen but not yet serialized) arc.
type BuilderArc[T any] struct {
	Label           int
	Target          Node
	IsFinal         bool
	Output          T
	NextFinalOutput T
}
