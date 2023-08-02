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

var _ Node = &UnCompiledNode{}

// UnCompiledNode
// TODO:
// instead of recording isFinal/output on the
// node, maybe we should use -1 arc to mean "end" (like
// we do when reading the FST).  Would simplify much
// code here...
type UnCompiledNode struct {
	Arcs       []*PendingArc
	Output     Output
	IsFinal    bool
	InputCount int
	Depth      int // This node's depth, starting from the automaton root.
	builder    *Builder
}

// PendingArc
// Expert: holds a pending (seen but not yet serialized) arc.
type PendingArc struct {
	Label           int
	Target          Node
	IsFinal         bool
	Output          Output
	NextFinalOutput Output
}

func (u *UnCompiledNode) NumArcs() int {
	return len(u.Arcs)
}

func NewUnCompiledNode(builder *Builder, depth int) *UnCompiledNode {
	return &UnCompiledNode{
		builder: builder,
		Arcs:    make([]*PendingArc, 0),
		Output:  builder.noOutput,
		Depth:   depth,
	}
}

func (u *UnCompiledNode) IsCompiled() bool {
	return false
}

func (u *UnCompiledNode) Code() int64 {
	return -1
}

func (u *UnCompiledNode) Clear() {
	u.Arcs = u.Arcs[:0]
	u.IsFinal = false
	u.Output = u.builder.noOutput
	u.InputCount = 0

	// We don't clear the depth here because it never changes
	// for nodes on the frontier (even when reused).
}

func (u *UnCompiledNode) GetLastOutput() Output {
	return u.lastArc().Output
}

func (u *UnCompiledNode) lastArc() *PendingArc {
	return u.Arcs[len(u.Arcs)-1]
}

func (u *UnCompiledNode) AddArc(label int, target Node) {
	u.Arcs = append(u.Arcs, &PendingArc{
		Label:           label,
		Target:          target,
		Output:          u.builder.noOutput,
		NextFinalOutput: u.builder.noOutput,
		IsFinal:         false,
	})
}

// DeleteLast 移除目标arc
func (u *UnCompiledNode) DeleteLast(ctx context.Context, label int, target Node) error {
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
func (u *UnCompiledNode) SetLastOutput(ctx context.Context, label int, newOutput Output) error {
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
func (u *UnCompiledNode) ReplaceLast(labelToMatch int, target Node, nextFinalOutput Output, isFinal bool) error {
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
func (u *UnCompiledNode) PrependOutput(outputPrefix Output) error {
	for i := range u.Arcs {
		output, err := outputPrefix.Add(u.Arcs[i].Output)
		if err != nil {
			return err
		}
		u.Arcs[i].Output = output
	}

	if u.IsFinal {
		output, err := outputPrefix.Add(u.Output)
		if err != nil {
			return err
		}
		u.Output = output
	}

	return nil
}
