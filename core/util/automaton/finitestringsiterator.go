package automaton

import (
	"errors"
	"github.com/bits-and-blooms/bitset"
)

// FiniteStringsIterator
// Iterates all accepted strings.
// If the Automaton has cycles then this iterator may throw an IllegalArgumentException, but this is not guaranteed!
// Be aware that the iteration order is implementation dependent and may change across releases.
// If the automaton is not determinized then it's possible this iterator will return duplicates.
type FiniteStringsIterator struct {
	// Automaton to create finite string from.
	a *Automaton

	// The state where each path should stop or -1 if only accepted states should be final.
	endState int

	// Tracks which states are in the current path, for cycle detection.
	pathStates *bitset.BitSet

	// Builder for current finite string.
	builder *IntsRefBuilder

	// Stack to hold our current state in the recursion/iteration.
	nodes []*PathNode

	// Emit empty string?.
	emitEmptyString bool
}

func NewFiniteStringsIterator(a *Automaton, startState, endState int) *FiniteStringsIterator {
	this := &FiniteStringsIterator{}

	this.a = a
	this.endState = endState
	this.nodes = make([]*PathNode, 16)
	for i := range this.nodes {
		this.nodes[i] = NewPathNode()
	}
	this.builder = NewIntsRefBuilder()
	this.pathStates = bitset.New(uint(a.GetNumStates()))
	this.emitEmptyString = a.IsAccept(0)

	// Start iteration with node startState.
	if a.GetNumTransitionsWithState(startState) > 0 {
		this.pathStates.Set(uint(startState))
		this.nodes[0].ResetState(a, startState)
		this.builder.Append(startState)
	}
	return this
}

var (
	EMPTYINTS = make([]int, 0)
)

func (f *FiniteStringsIterator) Next() ([]int, error) {
	if f.emitEmptyString {
		f.emitEmptyString = false
		return EMPTYINTS, nil
	}

	for depth := f.builder.Length(); depth > 0; {
		node := f.nodes[depth-1]

		// Get next label leaving the current node:
		label := node.NextLabel(f.a)
		if label != -1 {
			f.builder.Set(depth-1, label)

			to := node.to
			if f.a.GetNumTransitionsWithState(to) != 0 && to != f.endState {
				// Now recurse: the destination of this transition has outgoing transitions:
				if f.pathStates.Test(uint(to)) {
					return nil, errors.New("automaton has cycles")
				}
				f.pathStates.Set(uint(to))

				// Push node onto stack:
				f.growStack(depth)
				f.nodes[depth].ResetState(f.a, to)
				depth++
				f.builder.SetLength(depth)
				f.builder.Grow(depth)
			} else if f.endState == to || f.a.IsAccept(to) {
				// This transition leads to an accept state, so we save the current string:
				return f.builder.Get(), nil
			}
		} else {
			// No more transitions leaving this state, pop/return back to previous state:
			state := node.state
			f.pathStates.Clear(uint(state))
			depth--
			f.builder.SetLength(depth)

			if f.a.IsAccept(state) {
				// This transition leads to an accept state, so we save the current string:
				return f.builder.Get(), nil
			}
		}
	}
	return nil, nil
}

// Grow path stack, if required.
func (f *FiniteStringsIterator) growStack(depth int) {
	if len(f.nodes) == depth {
		f.nodes = append(f.nodes, NewPathNode())
	}
}

type PathNode struct {
	state      int
	to         int
	transition int
	label      int
	t          *Transition
}

func NewPathNode() *PathNode {
	return &PathNode{
		t: NewTransition(),
	}
}

func (p *PathNode) ResetState(a *Automaton, state int) {
	p.state = state
	p.transition = 0
	a.getTransition(state, 0, p.t)
	p.label = p.t.Min
	p.to = p.t.Dest
}

// NextLabel
// Returns next label of current transition, or advances to next transition and returns its
// first label, if current one is exhausted. If there are no more transitions, returns -1.
func (p *PathNode) NextLabel(a *Automaton) int {
	if p.label > p.t.Max {
		// We've exhaused the current transition's labels;
		// move to next transitions:
		p.transition++
		if p.transition >= a.GetNumTransitionsWithState(p.state) {
			// We're done iterating transitions leaving this state
			p.label = -1
			return -1
		}
		a.getTransition(p.state, p.transition, p.t)
		p.label = p.t.Min
		p.to = p.t.Dest
	}
	label := p.label
	p.label++
	return label
}
