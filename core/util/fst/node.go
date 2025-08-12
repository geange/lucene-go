package fst

import (
	"context"
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

func (r *CompiledNode) Code() int64 {
	return r.node
}

//var _ Node = &UnCompiledNode[]{}

// UnCompiledNode
// TODO:
// instead of recording isFinal/output on the
// node, maybe we should use -1 arc to mean "end" (like
// we do when reading the FST).  Would simplify much
// code here...
type UnCompiledNode[T any] struct {
	Arcs       []*PendingArc[T]
	Output     T
	IsFinal    bool
	InputCount int
	Depth      int // This node's depth, starting from the automaton root.
	builder    *Builder[T]
}

// PendingArc
// Expert: holds a pending (seen but not yet serialized) arc.
type PendingArc[T any] struct {
	Label           int
	Target          Node
	IsFinal         bool
	Output          T
	NextFinalOutput T
}

func (u *UnCompiledNode[T]) NumArcs() int {
	return len(u.Arcs)
}

func NewUnCompiledNode[T any](builder *Builder[T], depth int) *UnCompiledNode[T] {
	return &UnCompiledNode[T]{
		builder: builder,
		Arcs:    make([]*PendingArc[T], 0),
		Output:  builder.noOutput,
		Depth:   depth,
	}
}

func (u *UnCompiledNode[T]) IsCompiled() bool {
	return false
}

func (u *UnCompiledNode[T]) Code() int64 {
	return -1
}

func (u *UnCompiledNode[T]) Clear() {
	u.Arcs = u.Arcs[:0]
	u.IsFinal = false
	u.Output = u.builder.noOutput
	u.InputCount = 0

	// We don't clear the depth here because it never changes
	// for nodes on the frontier (even when reused).
}

func (u *UnCompiledNode[T]) GetLastOutput() T {
	return u.lastArc().Output
}

func (u *UnCompiledNode[T]) lastArc() *PendingArc[T] {
	return u.Arcs[len(u.Arcs)-1]
}

func (u *UnCompiledNode[T]) AddArc(label int, target Node) {
	u.Arcs = append(u.Arcs, &PendingArc[T]{
		Label:           label,
		Target:          target,
		Output:          u.builder.noOutput,
		NextFinalOutput: u.builder.noOutput,
		IsFinal:         false,
	})
}

// DeleteLast 移除目标arc
func (u *UnCompiledNode[T]) DeleteLast(ctx context.Context, label int, target Node) error {
	if len(u.Arcs) <= 0 {
		return errors.New("arcs size is 0")
	}

	lastArc := u.lastArc()

	if label != lastArc.Label {
		return errors.New("label not match")
	}

	if target != lastArc.Target {
		return errors.New("target not match")
	}
	u.Arcs = u.Arcs[:len(u.Arcs)-1]
	return nil
}

// SetLastOutput 设置最后arc的output对象
func (u *UnCompiledNode[T]) SetLastOutput(ctx context.Context, label int, newOutput T) error {
	if len(u.Arcs) <= 0 {
		return errors.New("arcs size is 0")
	}
	lastArc := u.lastArc()
	if lastArc.Label != label {
		return errors.New("label not match")
	}
	lastArc.Output = newOutput
	return nil
}

// ReplaceLast 替换最后的arc的内部数据
func (u *UnCompiledNode[T]) ReplaceLast(labelToMatch int, target Node, nextFinalOutput T, isFinal bool) error {
	if len(u.Arcs) <= 0 {
		return fmt.Errorf("arcs size is 0")
	}

	lastArc := u.lastArc()
	if lastArc.Label != labelToMatch {
		return fmt.Errorf("arc.label=%d vs %d", lastArc.Label, labelToMatch)
	}
	lastArc.Target = target
	lastArc.NextFinalOutput = nextFinalOutput
	lastArc.IsFinal = isFinal
	return nil
}

// PrependOutput pushes an output prefix forward onto all arcs
// 所有的边都增加一个output前缀
func (u *UnCompiledNode[T]) PrependOutput(outputPrefix T) error {
	for i, arc := range u.Arcs {
		output, err := u.builder.fst.outputs.Add(outputPrefix, arc.Output)
		if err != nil {
			return err
		}
		u.Arcs[i].Output = output
	}

	if u.IsFinal {
		output, err := u.builder.fst.outputs.Add(outputPrefix, u.Output)
		if err != nil {
			return err
		}
		u.Output = output
	}

	return nil
}
