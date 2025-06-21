package automaton

import (
	"fmt"
	"sort"

	"github.com/bits-and-blooms/bitset"
)

// Automaton Represents an automaton and all its states and transitions. States are integers and must be
// created using createState. Mark a state as an accept state using setAccept. Add transitions using
// addTransition. Each state must have all of its transitions added at once; if this is too restrictive
// then use Automaton.Builder instead. State 0 is always the initial state. Once a state is finished,
// either because you've starting adding transitions to another state or you call FinishState, then that
// states transitions are sorted (first by min, then max, then dest) and reduced (transitions with adjacent
// labels going to the same dest are combined).
type Automaton struct {
	// Where we next write to the int[] states; this increments by 2 for each added state because we
	// pack a pointer to the transitions array and a count of how many transitions leave the state.
	//nextState int

	// Where we next write to in int[] transitions; this increments by 3 for each added transition because
	// we pack min, max, dest in sequence.
	//nextTransition int

	// Current state we are adding transitions to; the caller must add all transitions for this state
	// before moving onto another state.
	curState int

	// Index in the transitions array, where this states leaving transitions are stored, or -1
	// if this state has not added any transitions yet, followed by number of transitions.
	states []int

	isAccept *bitset.BitSet

	// Holds toState, min, max for each transition.
	transitions []int

	// True if no state has two transitions leaving with the same label.
	deterministic bool
}

func NewAutomaton() *Automaton {
	return NewAutomatonV1(2, 2)
}

func NewAutomatonV1(numStates, numTransitions int) *Automaton {
	return &Automaton{
		curState:      -1,
		deterministic: true,
		states:        make([]int, 0, numStates*2),
		isAccept:      bitset.New(uint(numStates)),
		transitions:   make([]int, 0, numTransitions*3),
	}
}

// CreateState Create a new state.
func (a *Automaton) CreateState() int {
	state := len(a.states) / 2
	a.states = append(a.states, -1, 0)
	return state
	//state := a.nextState / 2
	//a.states[a.nextState] = -1
	//a.nextState += 2
	//return state
}

// SetAccept Set or clear this state as an accept state.
func (a *Automaton) SetAccept(state int, accept bool) {
	a.isAccept.SetTo(uint(state), accept)
}

// Sugar to get all transitions for all states. This is object-heavy; it's better to iterate state by state instead.
func (a *Automaton) getSortedTransitions() [][]Transition {
	numStates := a.GetNumStates()
	transitions := make([][]Transition, numStates)

	for s := 0; s < numStates; s++ {
		numTransitions := a.GetNumTransitionsWithState(s)
		transitions[s] = make([]Transition, numTransitions)

		for t := 0; t < numTransitions; t++ {
			transition := Transition{}
			a.getTransition(s, t, &transition)
			transitions[s][t] = transition
		}
	}

	return transitions
}

// Returns accept states. If the bit is set then that state is an accept state.
func (a *Automaton) getAcceptStates() *bitset.BitSet {
	return a.isAccept
}

// IsAccept Returns true if this state is an accept state.
func (a *Automaton) IsAccept(state int) bool {
	return a.isAccept.Test(uint(state))
}

// AddTransitionLabel Add a new transition with min = max = label.
func (a *Automaton) AddTransitionLabel(source, dest, label int) error {
	return a.AddTransition(source, dest, label, label)
}

// AddTransition Add a new transition with the specified source, dest, min, max.
func (a *Automaton) AddTransition(source, dest, min, max int) error {
	if a.curState != source {
		if a.curState != -1 {
			a.finishCurrentState()
		}

		// Move to next source:
		a.curState = source
		if a.states[2*a.curState] != -1 {
			return fmt.Errorf("from state (%d) already had transitions added", source)
		}
		a.states[2*a.curState] = len(a.transitions)
	}

	a.transitions = append(a.transitions, dest, min, max)

	//a.transitions[a.nextTransition] = dest
	//a.nextTransition++
	//a.transitions[a.nextTransition] = min
	//a.nextTransition++
	//a.transitions[a.nextTransition] = max
	//a.nextTransition++

	// Increment transition count for this state
	a.states[2*a.curState+1]++
	return nil
}

// AddEpsilon Add a [virtual] epsilon transition between source and dest. Dest state must already have all
// transitions added because this method simply copies those same transitions over to source.
func (a *Automaton) AddEpsilon(source, dest int) {
	t := Transition{}
	count := a.InitTransition(dest, &t)

	for i := 0; i < count; i++ {
		a.GetNextTransition(&t)
		_ = a.AddTransition(source, t.Dest, t.Min, t.Max)
	}

	if a.IsAccept(dest) {
		a.SetAccept(source, true)
	}
}

// Copy Copies over all states/transitions from other. The states numbers are sequentially assigned (appended).
func (a *Automaton) Copy(other *Automaton) {

	// Bulk copy and then fixup the state pointers:
	stateOffset := a.GetNumStates()

	nextTransition := len(a.transitions)
	nextState := len(a.states)

	a.states = append(a.states, other.states...)
	for i := nextState; i < len(a.states); i += 2 {
		a.states[i] += nextTransition
	}

	//a.nextState += other.nextState
	otherNumStates := other.GetNumStates()
	otherAcceptStates := other.getAcceptStates()
	state := uint(0)

	var ok bool
	for {
		if state < uint(otherNumStates) {
			if state, ok = otherAcceptStates.NextSet(state); ok {
				a.SetAccept(stateOffset+int(state), true)
				state++
				continue
			}
		}

		break
	}

	// Bulk copy and then fixup dest for each transition:
	//a.transitions = grow(a.transitions, a.nextTransition+other.nextTransition)
	//nextTransition := len(a.transitions)
	a.transitions = append(a.transitions, other.transitions...)
	//copy(a.transitions[a.nextTransition:a.nextTransition+other.nextTransition], other.transitions)
	for i := 0; i < len(other.transitions); i += 3 {
		a.transitions[nextTransition+i] += stateOffset
	}
	//a.nextTransition += other.nextTransition

	if other.deterministic == false {
		a.deterministic = false
	}
}

// Freezes the last state, sorting and reducing the transitions.
// 该函数finishCurrentState()的作用是整理当前状态的转移表，合并相邻区间并判断是否为确定性状态转移**。具体功能如下：
// 1. 排序转移项：根据目标状态和字符范围对转移进行排序；
// 2. 合并相邻区间：遍历所有转移，将相同目标状态且连续或相邻的字符区间合并；
// 3. 更新转移数量：减少冗余转移，更新实际有效转移数目；
// 4. 再次排序：按字符范围排序以提高匹配效率；
// 5. 检查确定性：若多个转移存在重叠输入范围，则标记为非确定性（deterministic = false）。
func (a *Automaton) finishCurrentState() {
	numTransitions := a.states[2*a.curState+1]
	offset := a.states[2*a.curState]

	start := offset / 3

	//根据目标状态和字符范围对转移进行排序
	sort.Sort(&destMinMaxSorter{
		Automaton: a,
		from:      start,
		to:        start + numTransitions,
	})

	// Reduce any "adjacent" transitions:
	upto := 0
	minValue := -1
	maxValue := -1
	dest := -1

	for i := 0; i < numTransitions; i++ {
		idx := offset + 3*i
		tDest := a.transitions[idx]
		tMin := a.transitions[idx+1]
		tMax := a.transitions[idx+2]

		if dest == tDest {
			if tMin <= maxValue+1 {
				if tMax > maxValue {
					maxValue = tMax
				}
			} else {
				if dest != -1 {
					uptoIdx := offset + 3*upto
					a.transitions[uptoIdx] = dest
					a.transitions[uptoIdx+1] = minValue
					a.transitions[uptoIdx+2] = maxValue
					upto++
				}
				minValue = tMin
				maxValue = tMax
			}
		} else {
			if dest != -1 {
				uptoIdx := offset + 3*upto
				a.transitions[uptoIdx] = dest
				a.transitions[uptoIdx+1] = minValue
				a.transitions[uptoIdx+2] = maxValue
				upto++
			}
			dest = tDest
			minValue = tMin
			maxValue = tMax
		}
	}

	if dest != -1 {
		// Last transition
		uptoIndex := offset + 3*upto
		a.transitions[uptoIndex] = dest
		a.transitions[uptoIndex+1] = minValue
		a.transitions[uptoIndex+2] = maxValue
		upto++
	}

	newTransitionsSize := len(a.transitions) - (numTransitions-upto)*3
	a.transitions = a.transitions[:newTransitionsSize]
	//
	//a.nextTransition -= (numTransitions - upto) * 3
	//a.states[2*a.curState+1] = upto

	// Sort transitions by minValue/maxValue/dest:
	sort.Sort(&minMaxDestSorter{
		Automaton: a,
		from:      start,
		to:        start + upto,
	})

	if a.deterministic && upto > 1 {
		lastMax := a.transitions[offset+2]
		for i := 1; i < upto; i++ {
			i3 := 3 * i
			minValue = a.transitions[offset+i3+1]
			if minValue <= lastMax {
				a.deterministic = false
				break
			}
			lastMax = a.transitions[offset+i3+2]
		}
	}
}

// IsDeterministic Returns true if this automaton is deterministic (for ever state there is only one
// transition for each label).
func (a *Automaton) IsDeterministic() bool {
	return a.deterministic
}

// FinishState
// Finishes the current state; call this once you are done adding transitions for a state.
// This is automatically called if you start adding transitions to a new source state,
// but for the last state you add you need to this method yourself.
func (a *Automaton) FinishState() {
	if a.curState != -1 {
		a.finishCurrentState()
		a.curState = -1
	}
}

// GetNumStates How many states this automaton has.
func (a *Automaton) GetNumStates() int {
	return len(a.states) / 2
}

// GetNumTransitions How many transitions this automaton has.
func (a *Automaton) GetNumTransitions() int {
	return len(a.transitions) / 3
}

// GetNumTransitionsWithState How many transitions this state has.
func (a *Automaton) GetNumTransitionsWithState(state int) int {
	idx := 2*state + 1
	if len(a.states) <= idx {
		return 0
	}
	count := a.states[idx]
	if count == -1 {
		return 0
	}
	return count
}

//func (a *Automaton) growStates() {
//	if a.nextState+2 > len(a.states) {
//		a.states = grow(a.states, a.nextState+2)
//	}
//}

//func (a *Automaton) growTransitions() {
//	if a.nextTransition+3 > len(a.transitions) {
//		a.transitions = grow(a.transitions, a.nextTransition+3)
//	}
//}

// Sorts transitions by dest, ascending, then min label ascending, then max label ascending
type destMinMaxSorter struct {
	*Automaton

	from, to int
}

func (r *destMinMaxSorter) Len() int {
	return r.to - r.from
}

func (r *destMinMaxSorter) Less(i, j int) bool {
	iStart := 3 * (i + r.from)
	jStart := 3 * (j + r.from)

	iDest := r.transitions[iStart]
	jDest := r.transitions[jStart]

	// First dest:
	if iDest < jDest {
		return true
	} else if iDest > jDest {
		return false
	}

	// Then min:
	iMin := r.transitions[iStart+1]
	jMin := r.transitions[jStart+1]
	if iMin < jMin {
		return true
	} else if iMin > jMin {
		return false
	}

	// Then max:
	iMax := r.transitions[iStart+2]
	jMax := r.transitions[jStart+2]
	if iMax < jMax {
		return true
	} else if iMax > jMax {
		return false
	}

	return false
}

func (r *destMinMaxSorter) Swap(i, j int) {
	iStart, jStart := 3*(i+r.from), 3*(j+r.from)
	r.swapOne(iStart, jStart)
	r.swapOne(iStart+1, jStart+1)
	r.swapOne(iStart+2, jStart+2)
}

func (r *destMinMaxSorter) swapOne(i, j int) {
	r.transitions[i], r.transitions[j] =
		r.transitions[j], r.transitions[i]
}

// Sorts transitions by min label, ascending, then max label ascending, then dest ascending
type minMaxDestSorter struct {
	from, to int
	*Automaton
}

func (r *minMaxDestSorter) Len() int {
	return r.to - r.from
}

func (r *minMaxDestSorter) Less(i, j int) bool {
	iStart := 3 * (i + r.from)
	jStart := 3 * (j + r.from)

	// First min:
	iMin := r.transitions[iStart+1]
	jMin := r.transitions[jStart+1]
	if iMin < jMin {
		return true
	} else if iMin > jMin {
		return false
	}

	// Then max:
	iMax := r.transitions[iStart+2]
	jMax := r.transitions[jStart+2]
	if iMax < jMax {
		return true
	} else if iMax > jMax {
		return false
	}

	// Then dest:
	iDest := r.transitions[iStart]
	jDest := r.transitions[jStart]
	if iDest < jDest {
		return true
	} else if iDest > jDest {
		return false
	}

	return false
}

func (r *minMaxDestSorter) Swap(i, j int) {
	iStart, jStart := 3*(i+r.from), 3*(j+r.from)
	r.swapOne(iStart, jStart)
	r.swapOne(iStart+1, jStart+1)
	r.swapOne(iStart+2, jStart+2)
}

func (r *minMaxDestSorter) swapOne(i, j int) {
	r.transitions[i], r.transitions[j] =
		r.transitions[j], r.transitions[i]
}

// InitTransition Initialize the provided Transition to iterate through all transitions leaving the specified
// state. You must call GetNextTransition to get each transition. Returns the number of transitions leaving
// this state.
func (a *Automaton) InitTransition(state int, t *Transition) int {
	t.Source = state
	t.TransitionUpto = a.states[2*state]
	return a.GetNumTransitionsWithState(state)
}

// GetNextTransition Iterate to the next transition after the provided one
func (a *Automaton) GetNextTransition(t *Transition) {
	t.Dest = a.transitions[t.TransitionUpto]
	t.TransitionUpto++
	t.Min = a.transitions[t.TransitionUpto]
	t.TransitionUpto++
	t.Max = a.transitions[t.TransitionUpto]
	t.TransitionUpto++
}

func (a *Automaton) transitionSorted(t *Transition) bool {
	upto := t.TransitionUpto
	if upto == a.states[2*t.Source] {
		// Transition isn't initialized yet (this is the first transition); don't check:
		return true
	}

	nextDest := a.transitions[upto]
	nextMin := a.transitions[upto+1]
	nextMax := a.transitions[upto+2]
	if nextMin > t.Min {
		return true
	} else if nextMin < t.Min {
		return false
	}

	// Min is equal, now test max:
	if nextMax > t.Max {
		return true
	} else if nextMax < t.Max {
		return false
	}

	// Max is also equal, now test dest:
	if nextDest > t.Dest {
		return true
	} else if nextDest < t.Dest {
		return false
	}

	// We should never see fully equal transitions here:
	return false
}

// Fill the provided Transition with the index'th transition leaving the specified state.
func (a *Automaton) getTransition(state, index int, t *Transition) {
	i := a.states[2*state] + 3*index
	t.Source = state
	t.Dest = a.transitions[i]
	i++
	t.Min = a.transitions[i]
	i++
	t.Max = a.transitions[i]
	i++
}

// GetStartPoints Returns sorted array of all interval start points.
func (a *Automaton) GetStartPoints() []int {
	pointset := make(map[int]struct{})
	pointset[0] = struct{}{}

	for s := 0; s < len(a.states); s += 2 {
		trans := a.states[s]
		limit := trans + 3*a.states[s+1]
		//System.out.println("  state=" + (s/2) + " trans=" + trans + " limit=" + limit);
		for trans < limit {
			minTrans := a.transitions[trans+1]
			maxTrans := a.transitions[trans+2]
			//System.out.println("    min=" + min);
			pointset[minTrans] = struct{}{}
			if maxTrans < 0x10FFFF {
				pointset[maxTrans+1] = struct{}{}
			}
			trans += 3
		}
	}

	points := make([]int, 0, len(pointset))
	for k, _ := range pointset {
		points = append(points, k)
	}
	sort.Ints(points)
	return points
}

// Step Performs lookup in transitions, assuming determinism.
// Params: 	state – starting state
//
//	label – codepoint to look up
//
// Returns: destination state, -1 if no matching outgoing transition
func (a *Automaton) Step(state, label int) int {
	return a.next(state, 0, label, nil)
}

// Next
// Looks for the next transition that matches the provided label, assuming determinism.
// This method is similar to step(int, int) but is used more efficiently when iterating over multiple
// transitions from the same source state. It keeps the latest reached transition index in
// transition.transitionUpto so the next call to this method can continue from there instead of restarting
// from the first transition.
//
// transition: The transition to start the lookup from (inclusive, using its Transition.source
// and Transition.transitionUpto). It is updated with the matched transition; or with
// Transition.dest = -1 if no match.
//
// label: The codepoint to look up.
//
// Returns: The destination state; or -1 if no matching outgoing transition.
func (a *Automaton) Next(transition *Transition, label int) int {
	return a.next(transition.Source, 0, label, transition)
}

// Looks for the next transition that matches the provided label, assuming determinism.
// state: The source state.
// fromTransitionIndex: The transition index to start the lookup from (inclusive); negative interpreted as 0.
// label: The codepoint to look up.
// transition: The output transition to update with the matching transition; or null for no update.
//
// Returns: The destination state; or -1 if no matching outgoing transition.
func (a *Automaton) next(state, fromTransitionIndex, label int, transition *Transition) int {
	stateIndex := 2 * state
	firstTransitionIndex := a.states[stateIndex]
	numTransitions := a.states[stateIndex+1]

	// Since transitions are sorted,
	// binary search the transition for which label is within [minLabel, maxLabel].
	low := max(fromTransitionIndex, 0)
	high := numTransitions - 1

	for low <= high {
		mid := (low + high) >> 1
		transitionIndex := firstTransitionIndex + 3*mid
		minLabel := a.transitions[transitionIndex+1]
		if minLabel > label {
			high = mid - 1
		} else {
			maxLabel := a.transitions[transitionIndex+2]
			if maxLabel < label {
				low = mid + 1
			} else {
				destState := a.transitions[transitionIndex]
				if transition != nil {
					transition.Dest = destState
					transition.Min = minLabel
					transition.Max = maxLabel
					transition.TransitionUpto = mid
				}
				return destState
			}
		}
	}

	destState := -1
	if transition != nil {
		transition.Dest = destState
		transition.TransitionUpto = low
	}
	return destState
}

var _ sort.Interface = &builderSorter{}

type builderSorter struct {
	values []int
	size   int
}

func (b *builderSorter) Len() int {
	return b.size
}

func (b *builderSorter) Less(i, j int) bool {
	i *= 4
	j *= 4

	if b.values[i] < b.values[j] {
		return true
	} else if b.values[i] > b.values[j] {
		return false
	}

	if b.values[i+1] < b.values[j+1] {
		return true
	} else if b.values[i+1] > b.values[j+1] {
		return false
	}

	if b.values[i+2] < b.values[j+2] {
		return true
	} else if b.values[i+2] > b.values[j+2] {
		return false
	}

	if b.values[i+3] < b.values[j+3] {
		return true
	} else if b.values[i+3] > b.values[j+3] {
		return false
	}

	return false
}

func (b *builderSorter) Swap(i, j int) {
	i *= 4
	j *= 4

	b.values[i], b.values[j] = b.values[j], b.values[i]
	b.values[i+1], b.values[j+1] = b.values[j+1], b.values[i+1]
	b.values[i+2], b.values[j+2] = b.values[j+2], b.values[i+2]
	b.values[i+3], b.values[j+3] = b.values[j+3], b.values[i+3]
}

func (r *Builder) sort(from, to int) {
	sort.Sort(&builderSorter{
		values: r.transitions,
		size:   to - from,
	})
}

func (r *Builder) IsAccept(state int) bool {
	return r.isAccept.Test(uint(state))
}
