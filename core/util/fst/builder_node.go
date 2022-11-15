package fst

import "errors"

type Node interface {
	IsCompiled() bool
}

var _ Node = &CompiledNode{}

type CompiledNode struct {
	node int64
}

func (*CompiledNode) IsCompiled() bool {
	return true
}

var _ Node = &UnCompiledNode{}

type UnCompiledNode struct {
	Owner   *Builder
	NumArcs int64

	// TODO: instead of recording isFinal/output on the
	// node, maybe we should use -1 arc to mean "end" (like
	// we do when reading the FST).  Would simplify much
	// code here...
	Arcs       []BuilderArc
	Output     any
	IsFinal    bool
	InputCount int64

	// This node's depth, starting from the automaton root.
	Depth int
}

func (u *UnCompiledNode) IsCompiled() bool {
	return false
}

func (u *UnCompiledNode) Clear() {
	u.NumArcs = 0
	u.IsFinal = false
	u.Output = u.Owner.NO_OUTPUT
	u.InputCount = 0

	// We don't clear the depth here because it never changes
	// for nodes on the frontier (even when reused).
}

func (u *UnCompiledNode) GetLastOutput(labelToMatch int) any {
	return u.Arcs[len(u.Arcs)-1].Output
}

func (u *UnCompiledNode) AddArc(label int, target Node) {
	u.Arcs = append(u.Arcs, BuilderArc{
		Label:           label,
		Target:          target,
		Output:          nil,
		NextFinalOutput: nil,
		IsFinal:         false,
	})
}

func (u *UnCompiledNode) DeleteLast(label int, target Node) error {
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

func (u *UnCompiledNode) SetLastOutput(label int, newOutput any) error {
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

func (u *UnCompiledNode) ReplaceLast(target Node, nextFinalOutput any, isFinal bool) {
	arc := u.Arcs[len(u.Arcs)-1]
	arc.Target = target
	arc.NextFinalOutput = nextFinalOutput
	arc.IsFinal = isFinal
}

// PrependOutput pushes an output prefix forward onto all arcs
func (u *UnCompiledNode) PrependOutput(outputPrefix any) error {
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
type BuilderArc struct {
	Label           int
	Target          Node
	IsFinal         bool
	Output          any
	NextFinalOutput any
}
