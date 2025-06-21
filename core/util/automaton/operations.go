package automaton

import (
	"bytes"
	"cmp"
	"errors"
	"slices"
	"sync/atomic"
	"unicode"

	"github.com/bits-and-blooms/bitset"
)

const (
	DEFAULT_DETERMINIZE_WORK_LIMIT = 10000
)

// DeterminizeAutomaton Determinizes the given automaton.
// Worst case complexity: exponential in number of states.
// Params: 	workLimit – Maximum amount of "work" that the powerset construction will spend before throwing
//
//	TooComplexToDeterminizeException. Higher numbers allow this operation to consume more memory and
//	CPU but allow more complex automatons. Use DEFAULT_DETERMINIZE_WORK_LIMIT as a decent default
//	if you don't otherwise know what to specify.
//
// Throws: TooComplexToDeterminizeException – if determinizing requires more than workLimit "effort"
func DeterminizeAutomaton(a *Automaton, workLimit int) *Automaton {
	if a.IsDeterministic() {
		return a
	}
	if a.GetNumStates() <= 1 {
		// Already determinized
		return a
	}

	// subset construction
	b := NewBuilder()

	// Same initial values and state will always have the same hashCode
	initialSet := NewFrozenIntSet([]int{0}, uint64(mix32(0)+1), 0)
	// Create state 0:
	b.CreateState()

	worklist := make([]*FrozenIntSet, 0)
	newState := NewHashMap[int](WithCapacity(1))

	worklist = append(worklist, initialSet)
	b.SetAccept(0, a.IsAccept(0))
	newState.Set(initialSet, 0)

	// TODO:

	return a
}

// IsEmptyAutomaton
// Returns true if the given automaton accepts no strings.
func IsEmptyAutomaton(a *Automaton) bool {
	if a.GetNumStates() == 0 {
		// Common case: no states
		return true
	}

	if a.IsAccept(0) == false && a.GetNumTransitionsWithState(0) == 0 {
		// Common case: just one initial state
		return true
	}
	if a.IsAccept(0) == true {
		// Apparently common case: it accepts the damned empty string
		return false
	}

	workList := make([]int, 0)
	seen := bitset.New(uint(a.GetNumStates()))
	workList = append(workList, 0)
	seen.Set(0)

	t := NewTransition()
	for len(workList) > 0 {
		state := workList[0]
		workList = workList[1:]

		if a.IsAccept(state) {
			return false
		}

		count := a.InitTransition(state, t)
		for i := 0; i < count; i++ {
			a.GetNextTransition(t)
			if seen.Test(uint(t.Dest)) == false {
				workList = append(workList, t.Dest)
				seen.Set(uint(t.Dest))
			}
		}
	}
	return true
}

// IsTotalAutomaton
// Returns true if the given automaton accepts all strings. The automaton must be minimized.
func IsTotalAutomaton(a *Automaton) bool {
	return IsTotalAutomatonRange(a, 0, 0x10FFFF)
}

// IsTotalAutomatonRange
// Returns true if the given automaton accepts all strings for the specified min/max range of the alphabet.
// The automaton must be minimized.
func IsTotalAutomatonRange(a *Automaton, minAlphabet, maxAlphabet int) bool {
	if a.IsAccept(0) && a.GetNumTransitionsWithState(0) == 1 {
		t := NewTransition()
		a.getTransition(0, 0, t)
		return t.Dest == 0 && t.Min == minAlphabet && t.Max == maxAlphabet
	}
	return false
}

func GetSingletonAutomaton(a *Automaton) ([]int, error) {
	if a.IsDeterministic() == false {
		return nil, errors.New("input automaton must be deterministic")
	}

	ints := make([]int, 0)
	visited := make(map[int]struct{})
	s := 0
	t := NewTransition()
	for {
		visited[s] = struct{}{}

		if a.IsAccept(s) == false {
			if a.GetNumTransitionsWithState(s) == 1 {
				a.getTransition(s, 0, t)
				if _, ok := visited[t.Dest]; t.Min == t.Max && ok {
					ints = append(ints, t.Min)
					s = t.Dest
					continue
				}
			}
		} else if a.GetNumTransitionsWithState(s) == 0 {
			return ints, nil
		}

		// Automaton accepts more than one string:
		return nil, nil
	}
}

func IsFiniteAutomaton(a *Automaton) *atomic.Bool {
	flag := &atomic.Bool{}

	if a.GetNumStates() == 0 {
		flag.Store(true)
		return flag
	}

	b1 := bitset.New(uint(a.GetNumStates()))
	b2 := bitset.New(uint(a.GetNumStates()))

	return isFinite(NewTransition(), a, 0, b1, b2, 0)
}

// Checks whether there is a loop containing state. (This is sufficient since there are never transitions to dead states.)
// TODO: not great that this is recursive... in theory a
// large automata could exceed java's stack so the maximum level of recursion is bounded to 1000
func isFinite(scratch *Transition, a *Automaton, state int, path, visited *bitset.BitSet, level int) *atomic.Bool {
	flag := &atomic.Bool{}

	// if (level > MAX_RECURSION_LEVEL) {
	//      throw new IllegalArgumentException("input automaton is too large: " +  level);
	//    }
	path.Set(uint(state))
	numTransitions := a.InitTransition(state, scratch)
	for t := 0; t < numTransitions; t++ {
		a.getTransition(state, t, scratch)
		if path.Test(uint(scratch.Dest)) || (!visited.Test(uint(scratch.Dest)) && !isFinite(scratch, a, scratch.Dest, path, visited, level+1).Load()) {
			flag.Store(false)
			return flag
		}
	}
	path.Clear(uint(state))
	visited.Set(uint(state))
	flag.Store(true)
	return flag
}

// getCommonSuffixBytesRef
// Returns the longest BytesRef that is a suffix of all accepted strings. Worst case complexity: quadratic with the number of states+transitions.
// Returns: common suffix, which can be an empty (length 0) BytesRef (never null)
func getCommonSuffixBytesRef(a *Automaton) ([]byte, error) {
	// reverse the language of the automaton, then reverse its common prefix.
	ra, err := reverse(a)
	if err != nil {
		return nil, err
	}
	r, err := removeDeadStates(ra)
	if err != nil {
		return nil, err
	}

	ref, err := getCommonPrefixBytesRef(r)
	if err != nil {
		return nil, err
	}
	slices.Reverse(ref)
	return ref, nil
}

// Returns true if there are dead states reachable from an initial state.
func hasDeadStatesFromInitial(a *Automaton) bool {
	reachableFromInitial := getLiveStatesFromInitial(a)
	reachableFromAccept := getLiveStatesToAccept(a)
	reachableFromInitial = reachableFromInitial.Difference(reachableFromAccept)
	return reachableFromInitial.Count() > 0
}

func getCommonPrefix(a *Automaton) (string, error) {

	if hasDeadStatesFromInitial(a) {
		return "", errors.New("input automaton has dead states")
	}
	if isEmpty(a) {
		return "", nil
	}
	builder := new(bytes.Buffer)
	scratch := NewTransition()
	visited := bitset.New(uint(a.GetNumStates()))
	current := bitset.New(uint(a.GetNumStates()))
	next := bitset.New(uint(a.GetNumStates()))
	current.Set(0) // start with initial state
OUT:
	for {
		label := -1
		// do a pass, stepping all current paths forward once
		state, ok := current.NextSet(0)
		for ok {
			visited.Set(state)
			// if it is an accept state, we are done
			if a.IsAccept(int(state)) {
				break OUT
			}
			for transition := 0; transition < a.GetNumTransitionsWithState(int(state)); transition++ {
				a.getTransition(int(state), transition, scratch)
				if label == -1 {
					label = scratch.Min
				}
				// either a range of labels, or label that doesn't match all the other paths this round
				if scratch.Min != scratch.Max || scratch.Min != label {
					break OUT
				}
				// mark target state for next iteration
				next.Set(uint(scratch.Dest))
			}

			if state+1 >= current.Len() {
				ok = false
			} else {
				state, ok = current.NextSet(state + 1)
			}
		}

		// add the label to the prefix
		builder.WriteRune(rune(label))
		// swap "current" with "next", clear "next"
		tmp := current
		current = next
		next = tmp
		next.ClearAll()
	}
	return builder.String(), nil
}

func isEmpty(a *Automaton) bool {
	if a.GetNumStates() == 0 {
		// Common case: no states
		return true
	}
	if a.IsAccept(0) == false && a.GetNumTransitionsWithState(0) == 0 {
		// Common case: just one initial state
		return true
	}
	if a.IsAccept(0) == true {
		// Apparently common case: it accepts the damned empty string
		return false
	}

	workList := make([]int, 0)

	seen := bitset.New(uint(a.GetNumStates()))
	workList = append(workList, 0)
	seen.Set(0)

	t := NewTransition()
	for len(workList) > 0 {
		state := workList[0]
		workList = workList[1:]
		if a.IsAccept(state) {
			return false
		}
		count := a.InitTransition(state, t)
		for i := 0; i < count; i++ {
			a.GetNextTransition(t)
			if seen.Test(uint(t.Dest)) == false {
				workList = append(workList, t.Dest)
				seen.Set(uint(t.Dest))
			}
		}
	}

	return true
}

func getCommonPrefixBytesRef(a *Automaton) ([]byte, error) {
	prefix, err := getCommonPrefix(a)
	if err != nil {
		return nil, err
	}
	builder := new(bytes.Buffer)

	for _, ch := range prefix {
		if ch > 255 {
			return nil, errors.New("automaton is not binary")
		}
		builder.WriteRune(ch)
	}

	return builder.Bytes(), nil
}

func reverse(a *Automaton) (*Automaton, error) {
	return reverseStates(a, nil)
}

func reverseStates(a *Automaton, initialStates map[int]struct{}) (*Automaton, error) {

	if isEmpty(a) {
		return NewAutomaton(), nil
	}

	numStates := a.GetNumStates()

	// Build a new automaton with all edges reversed
	builder := NewBuilder()

	// Initial node; we'll add epsilon transitions in the end:
	builder.CreateState()

	for s := 0; s < numStates; s++ {
		builder.CreateState()
	}

	// Old initial state becomes new accept state:
	builder.SetAccept(1, true)

	t := NewTransition()
	for s := 0; s < numStates; s++ {
		numTransitions := a.GetNumTransitionsWithState(s)
		a.InitTransition(s, t)
		for i := 0; i < numTransitions; i++ {
			a.GetNextTransition(t)
			builder.AddTransition(t.Dest+1, s+1, t.Min, t.Max)
		}
	}

	result := builder.Finish()

	s := uint(0)
	var ok bool
	acceptStates := a.getAcceptStates()
	for int(s) < numStates {
		s, ok = acceptStates.NextSet(s)
		if !ok {
			break
		}

		result.AddEpsilon(0, int(s+1))
		if initialStates != nil {
			initialStates[int(s+1)] = struct{}{}
		}
		s++
	}

	result.FinishState()

	return result, nil
}

func reverseAutomaton(a *Automaton) *Automaton {
	return reverseAutomatonIntSet(a, nil)
}

func removeDeadStates(a *Automaton) (*Automaton, error) {
	numStates := a.GetNumStates()
	liveSet := getLiveStates(a)

	mp := make([]int, numStates)

	result := NewAutomaton()
	for i := 0; i < numStates; i++ {
		if liveSet.Test(uint(i)) {
			mp[i] = result.CreateState()
			result.SetAccept(mp[i], a.IsAccept(i))
		}
	}

	t := NewTransition()

	for i := 0; i < numStates; i++ {
		if liveSet.Test(uint(i)) {
			numTransitions := a.InitTransition(i, t)
			// filter out transitions to dead states:
			for j := 0; j < numTransitions; j++ {
				a.GetNextTransition(t)
				if liveSet.Test(uint(t.Dest)) {
					err := result.AddTransition(mp[i], mp[t.Dest], t.Min, t.Max)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	result.FinishState()
	//assert hasDeadStates(result) == false;
	return result, nil
}

func getLiveStates(a *Automaton) *bitset.BitSet {
	live := getLiveStatesFromInitial(a)
	live.Union(getLiveStatesToAccept(a))
	return live
}

func getLiveStatesFromInitial(a *Automaton) *bitset.BitSet {
	numStates := a.GetNumStates()
	live := bitset.New(uint(numStates))
	if numStates == 0 {
		return live
	}
	workList := make([]int, 0)
	live.Set(0)
	workList = append(workList, 0)

	t := NewTransition()
	for len(workList) > 0 {
		s := workList[0]
		workList = workList[1:]
		count := a.InitTransition(s, t)
		for i := 0; i < count; i++ {
			a.GetNextTransition(t)
			if live.Test(uint(t.Dest)) == false {
				live.Set(uint(t.Dest))
				workList = append(workList, t.Dest)
			}
		}
	}

	return live
}

func getLiveStatesToAccept(a *Automaton) *bitset.BitSet {
	builder := NewBuilder()

	// NOTE: not quite the same thing as what SpecialOperations.reverse does:
	t := NewTransition()
	numStates := a.GetNumStates()
	for s := 0; s < numStates; s++ {
		builder.CreateState()
	}
	for s := 0; s < numStates; s++ {
		count := a.InitTransition(s, t)
		for i := 0; i < count; i++ {
			a.GetNextTransition(t)
			builder.AddTransition(t.Dest, s, t.Min, t.Max)
		}
	}
	a2 := builder.Finish()

	workList := make([]int, 0)
	live := bitset.New(uint(numStates))
	acceptBits := a.getAcceptStates()
	s := uint(0)
	ok := false
	for int(s) < numStates {
		s, ok = acceptBits.NextSet(s)
		if !ok {
			break
		}

		live.Set(s)
		workList = append(workList, int(s))
		s++
	}

	for len(workList) > 0 {
		state := workList[0]
		workList = workList[1:]
		count := a2.InitTransition(state, t)
		for i := 0; i < count; i++ {
			a2.GetNextTransition(t)
			if live.Test(uint(t.Dest)) == false {
				live.Set(uint(t.Dest))
				workList = append(workList, t.Dest)
			}
		}
	}

	return live
}

func reverseAutomatonIntSet(a *Automaton, initialStates map[int]struct{}) *Automaton {
	if IsEmptyAutomaton(a) {
		return NewAutomaton()
	}

	numStates := a.GetNumStates()

	// Build a new automaton with all edges reversed
	builder := NewBuilder()

	// Initial node; we'll add epsilon transitions in the end:
	builder.CreateState()

	for s := 0; s < numStates; s++ {
		builder.CreateState()
	}

	// Old initial state becomes new accept state:
	builder.SetAccept(1, true)

	t := NewTransition()
	for s := 0; s < numStates; s++ {
		numTransitions := a.GetNumTransitionsWithState(s)
		a.InitTransition(s, t)
		for i := 0; i < numTransitions; i++ {
			a.GetNextTransition(t)
			builder.AddTransition(t.Dest+1, s+1, t.Min, t.Max)
		}
	}

	result := builder.Finish()

	s := 0
	acceptStates := a.getAcceptStates()
	for {
		if _, ok := acceptStates.NextSet(uint(s)); !(ok && s < numStates) {
			break
		}

		result.AddEpsilon(0, s+1)
		if initialStates != nil {
			initialStates[s+1] = struct{}{}
		}
		s++
	}

	result.FinishState()

	return result
}

func union(automatons ...*Automaton) (*Automaton, error) {
	result := NewAutomaton()

	// Create initial state:
	result.CreateState()

	// Copy over all automata
	for _, a := range automatons {
		result.Copy(a)
	}

	// Add epsilon transition from new initial state
	stateOffset := 1
	for _, a := range automatons {
		if a.GetNumStates() == 0 {
			continue
		}
		result.AddEpsilon(0, stateOffset)
		stateOffset += a.GetNumStates()
	}

	result.FinishState()

	return removeDeadStates(result)
}

func concatenate(automatons ...*Automaton) (*Automaton, error) {
	result := NewAutomaton()

	// First pass: create all states
	for _, a := range automatons {
		if a.GetNumStates() == 0 {
			result.FinishState()
			return result, nil
		}
		numStates := a.GetNumStates()
		for s := 0; s < numStates; s++ {
			result.CreateState()
		}
	}

	// Second pass: add transitions, carefully linking accept
	// states of A to init state of next A:
	stateOffset := 0
	t := NewTransition()

	for i, a := range automatons {
		numStates := a.GetNumStates()

		var nextA *Automaton
		nextIdx := i + 1
		if nextIdx < len(automatons) {
			nextA = automatons[nextIdx]
		}

		for s := 0; s < numStates; s++ {
			numTransitions := a.InitTransition(s, t)
			for j := 0; j < numTransitions; j++ {
				a.GetNextTransition(t)

				srcState := stateOffset + s
				destState := stateOffset + t.Dest
				err := result.AddTransition(srcState, destState, t.Min, t.Max)
				if err != nil {
					return nil, err
				}
			}

			if a.IsAccept(s) {
				followA := nextA
				followOffset := stateOffset
				upto := i + 1
				for {
					if followA == nil {
						result.SetAccept(stateOffset+s, true)
						break
					}

					// Adds a "virtual" epsilon transition:
					numTransitions = followA.InitTransition(0, t)
					for j := 0; j < numTransitions; j++ {
						followA.GetNextTransition(t)
						err := result.AddTransition(stateOffset+s, followOffset+numStates+t.Dest, t.Min, t.Max)
						if err != nil {
							return nil, err
						}
					}
					if followA.IsAccept(0) {
						// Keep chaining if followA accepts empty string
						followOffset += followA.GetNumStates()
						if upto == len(automatons)-1 {
							followA = nil
						} else {
							followA = automatons[upto+1]
						}
						upto++
					} else {
						break
					}
				}
			}
		}

		stateOffset += numStates
	}

	if result.GetNumStates() == 0 {
		result.CreateState()
	}

	result.FinishState()

	return result, nil
}

func totalize(a *Automaton) (*Automaton, error) {
	result := NewAutomaton()
	numStates := a.GetNumStates()
	for i := 0; i < numStates; i++ {
		result.CreateState()
		result.SetAccept(i, a.IsAccept(i))
	}

	deadState := result.CreateState()
	err := result.AddTransition(deadState, deadState, 0, unicode.MaxRune)
	if err != nil {
		return nil, err
	}

	t := NewTransition()
	for i := 0; i < numStates; i++ {
		maxi := 0
		count := a.InitTransition(i, t)
		for j := 0; j < count; j++ {
			a.GetNextTransition(t)
			err := result.AddTransition(i, t.Dest, t.Min, t.Max)
			if err != nil {
				return nil, err
			}
			if t.Min > maxi {
				err := result.AddTransition(i, deadState, maxi, t.Min-1)
				if err != nil {
					return nil, err
				}
			}
			if t.Max+1 > maxi {
				maxi = t.Max + 1
			}
		}

		if maxi <= unicode.MaxRune {
			err := result.AddTransition(i, deadState, maxi, unicode.MaxRune)
			if err != nil {
				return nil, err
			}
		}
	}

	result.FinishState()
	return result, nil
}

func complement(a *Automaton, determinizeWorkLimit int) (*Automaton, error) {
	a, err := determinize(a, determinizeWorkLimit)
	if err != nil {
		return nil, err
	}
	a, err = totalize(a)
	if err != nil {
		return nil, err
	}
	numStates := a.GetNumStates()
	for p := 0; p < numStates; p++ {
		a.SetAccept(p, !a.IsAccept(p))
	}
	return removeDeadStates(a)
}

func determinize(a *Automaton, workLimit int) (*Automaton, error) {
	if a.IsDeterministic() {
		// Already determinized
		return a, nil
	}
	if a.GetNumStates() <= 1 {
		// Already determinized
		return a, nil
	}

	// subset construction
	b := NewBuilder()

	//System.out.println("DET:");
	//a.writeDot("/l/la/lucene/core/detin.dot");

	// Same initial values and state will always have the same hashCode
	initialset := NewFrozenIntSet([]int{0}, uint64(mix(0)+1), 0)

	// Create state 0:
	b.CreateState()

	worklist := make([]*FrozenIntSet, 0)
	newstate := NewHashMap[int]()

	worklist = append(worklist, initialset)

	b.SetAccept(0, a.IsAccept(0))
	newstate.Set(initialset, 0)

	// like Set<Integer,PointTransitions>
	points := NewPointTransitionSet()

	// like HashMap<Integer,Integer>, maps state to its count
	statesSet := NewStateSet()

	t := NewTransition()

	effortSpent := 0

	// LUCENE-9981: approximate conversion from what used to be a limit on number of states, to
	// maximum "effort":
	effortLimit := workLimit * 10

	for len(worklist) > 0 {
		// TODO (LUCENE-9983): these int sets really do not need to be sorted, and we are paying
		// a high (unecessary) price for that!  really we just need a low-overhead Map<int,int>
		// that implements equals/hash based only on the keys (ignores the values).  fixing this
		// might be a bigspeedup for determinizing complex automata
		s := worklist[0]
		worklist = worklist[1:]

		// LUCENE-9981: we more carefully aggregate the net work this automaton is costing us, instead
		// of (overly simplistically) counting number
		// of determinized states:
		effortSpent += len(s.values)
		if effortSpent >= effortLimit {
			return nil, errors.New("too Complex To Determinize")
		}

		// Collate all outgoing transitions by min/1+max:
		for i := 0; i < len(s.values); i++ {
			s0 := s.values[i]
			numTransitions := a.GetNumTransitionsWithState(s0)
			a.InitTransition(s0, t)
			for j := 0; j < numTransitions; j++ {
				a.GetNextTransition(t)
				points.Add(t)
			}
		}

		if len(points.points) == 0 {
			// No outgoing transitions -- skip it
			continue
		}

		points.Sort()

		lastPoint := -1
		accCount := 0

		r := s.state

		for i := 0; i < len(points.points); i++ {

			point := points.points[i].point

			if statesSet.Size() > 0 {

				q, ok := newstate.Get(statesSet)
				if !ok {
					q = b.CreateState()
					p := statesSet.Freeze(q)
					//System.out.println("  make new state=" + q + " -> " + p + " accCount=" + accCount);
					worklist = append(worklist, p)
					b.SetAccept(q, accCount > 0)
					newstate.Set(p, q)
				}

				// System.out.println("  add trans src=" + r + " dest=" + q + " min=" + lastPoint + " max=" + (point-1));

				b.AddTransition(r, q, lastPoint, point-1)
			}

			// process transitions that end on this point
			// (closes an overlapping interval)
			transitions := points.points[i].ends.transitions
			limit := points.points[i].ends.next
			for j := 0; j < limit; j += 3 {
				dest := transitions[j]
				statesSet.Decr(dest)
				if a.IsAccept(dest) {
					accCount--
				}
				//accCount -= a.isAccept(dest) ? 1:0;
			}
			points.points[i].ends.next = 0

			// process transitions that start on this point
			// (opens a new interval)
			transitions = points.points[i].starts.transitions
			limit = points.points[i].starts.next
			for j := 0; j < limit; j += 3 {
				dest := transitions[j]
				statesSet.Incr(dest)
				if a.IsAccept(dest) {
					accCount++
				}
			}
			lastPoint = point
			points.points[i].starts.next = 0
		}
		points.Reset()
	}

	result := b.Finish()
	return result, nil
}

type TransitionList struct {
	transitions []int
	next        int
}

func (t *TransitionList) reset() {
	t.next = 0
	t.transitions = t.transitions[:0]
}

func NewTransitionList() *TransitionList {
	return &TransitionList{
		transitions: make([]int, 0),
	}
}

func (t *TransitionList) Add(item *Transition) {
	t.transitions = append(t.transitions, item.Dest, item.Min, item.Max)
	t.next += 3
}

type PointTransitions struct {
	point  int
	ends   *TransitionList
	starts *TransitionList
}

func NewPointTransitions() *PointTransitions {
	return &PointTransitions{
		starts: NewTransitionList(),
		ends:   NewTransitionList(),
	}
}

func (p *PointTransitions) reset(point int) {
	p.point = point
	p.starts.reset()
	p.ends.reset()
}

type PointTransitionSet struct {
	points []*PointTransitions
	imap   map[int]*PointTransitions
}

func (s *PointTransitionSet) find(point int) *PointTransitions {
	p, ok := s.imap[point]
	if !ok {
		p = s.next(point)
		s.imap[point] = p
	}
	return p
}

func (s *PointTransitionSet) next(point int) *PointTransitions {
	points0 := NewPointTransitions()
	s.points = append(s.points, points0)
	points0.reset(point)
	return points0
}

func (s *PointTransitionSet) Add(t *Transition) {
	s.find(t.Min).starts.Add(t)
	s.find(1 + t.Max).ends.Add(t)
}

func (s *PointTransitionSet) Sort() {
	slices.SortStableFunc(s.points, func(e, e2 *PointTransitions) int {
		return cmp.Compare(e.point, e2.point)
		//if e.point < e2.point {
		//	return -1
		//} else if e.point == e2.point {
		//	return 0
		//}
		//return 1
	})
}

func (s *PointTransitionSet) Reset() {
	clear(s.imap)
	s.points = s.points[:0]
}

func NewPointTransitionSet() *PointTransitionSet {
	return &PointTransitionSet{
		points: make([]*PointTransitions, 0),
		imap:   make(map[int]*PointTransitions),
	}
}

func repeat(a *Automaton) (*Automaton, error) {
	if a.GetNumStates() == 0 {
		// Repeating the empty automata will still only accept the empty automata.
		return a, nil
	}
	builder := NewBuilder()
	builder.CreateState()
	builder.SetAccept(0, true)
	builder.CopyStates(a)

	t := NewTransition()
	count := a.InitTransition(0, t)
	for i := 0; i < count; i++ {
		a.GetNextTransition(t)
		builder.AddTransition(0, t.Dest+1, t.Min, t.Max)
	}

	numStates := a.GetNumStates()
	for s := 0; s < numStates; s++ {
		if a.IsAccept(s) {
			count = a.InitTransition(0, t)
			for i := 0; i < count; i++ {
				a.GetNextTransition(t)
				builder.AddTransition(s+1, t.Dest+1, t.Min, t.Max)
			}
		}
	}

	return builder.Finish(), nil
}

func repeatCount(a *Automaton, count int) (*Automaton, error) {
	if count == 0 {
		return repeat(a)
	}
	as := make([]*Automaton, 0)
	for count > 0 {
		count--
		as = append(as, a)
	}

	ra, err := repeat(a)
	if err != nil {
		return nil, err
	}
	as = append(as, ra)

	return concatenate(as...)
}

func repeatRange(a *Automaton, min, max int) (*Automaton, error) {
	if min > max {
		return defaultAutomata.MakeEmpty(), nil
	}

	var b *Automaton
	var err error
	if min == 0 {
		b = defaultAutomata.MakeEmptyString()
	} else if min == 1 {
		b = NewAutomaton()
		b.Copy(a)
	} else {
		as := make([]*Automaton, 0)
		for i := 0; i < min; i++ {
			as = append(as, a)
		}
		b, err = concatenate(as...)
		if err != nil {
			return nil, err
		}
	}

	prevAcceptStates := toSet(b, 0)
	builder := NewBuilder()
	builder.Copy(b)
	for i := min; i < max; i++ {
		numStates := builder.GetNumStates()
		builder.Copy(a)
		for s := range prevAcceptStates {
			builder.AddEpsilon(s, numStates)
		}
		prevAcceptStates = toSet(a, numStates)
	}

	return builder.Finish(), nil
}

func toSet(a *Automaton, offset int) map[int]struct{} {
	numStates := uint(a.GetNumStates())
	isAccept := a.getAcceptStates()
	result := make(map[int]struct{})
	upto := uint(0)
	var ok bool
	for upto < numStates {
		upto, ok = isAccept.NextSet(upto)
		if !ok {
			break
		}
		result[offset+int(upto)] = struct{}{}
		upto++
	}

	return result
}

var _ Hashable = &statePair{}

type statePair struct {
	s  int
	s1 int
	s2 int
}

func newStatePair(s int, s1 int, s2 int) *statePair {
	return &statePair{s: s, s1: s1, s2: s2}
}

func (s *statePair) Hash() uint64 {
	return uint64(s.s1*31 + s.s2)
}

func (s *statePair) Equals(other Hashable) bool {
	sp, ok := other.(*statePair)
	if !ok {
		return false
	}
	return s.s1 == sp.s1 && s.s2 == sp.s2
}

func intersection(a1, a2 *Automaton) (*Automaton, error) {
	if a1 == a2 {
		return a1, nil
	}
	if a1.GetNumStates() == 0 {
		return a1, nil
	}
	if a2.GetNumStates() == 0 {
		return a2, nil
	}
	transitions1 := a1.getSortedTransitions()
	transitions2 := a2.getSortedTransitions()
	c := NewAutomaton()
	c.CreateState()
	worklist := make([]*statePair, 0)
	estates := NewHashMap[*statePair]()

	p := newStatePair(0, 0, 0)
	worklist = append(worklist, p)
	estates.Set(p, p)
	for len(worklist) > 0 {
		p = worklist[0]
		worklist = worklist[1:]
		c.SetAccept(p.s, a1.IsAccept(p.s1) && a2.IsAccept(p.s2))
		t1 := transitions1[p.s1]
		t2 := transitions2[p.s2]
		n1 := 0
		b2 := 0
		for ; n1 < len(t1); n1++ {
			for b2 < len(t2) && t2[b2].Max < t1[n1].Min {
				b2++
			}

			n2 := b2
			for ; n2 < len(t2) && t1[n1].Max >= t2[n2].Min; n2++ {

			}
			if t2[n2].Max >= t1[n1].Min {
				q := newStatePair(-1, t1[n1].Dest, t2[n2].Dest)
				r, ok := estates.Get(q)
				if !ok {
					q.s = c.CreateState()
					worklist = append(worklist, q)
					estates.Set(q, q)
					r = q
				}
				var minI, maxI int

				if t1[n1].Min > t2[n2].Min {
					minI = t1[n1].Min
				} else {
					minI = t2[n2].Min
				}

				if t1[n1].Max < t2[n2].Max {
					maxI = t1[n1].Max
				} else {
					maxI = t2[n2].Max
				}

				err := c.AddTransition(p.s, r.s, minI, maxI)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	c.FinishState()

	return removeDeadStates(c)
}

func optional(a *Automaton) (*Automaton, error) {
	result := NewAutomaton()
	result.CreateState()
	result.SetAccept(0, true)
	if a.GetNumStates() > 0 {
		result.Copy(a)
		result.AddEpsilon(0, 1)
	}
	result.FinishState()
	return result, nil
}
