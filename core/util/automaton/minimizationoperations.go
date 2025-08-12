package automaton

import (
	"unicode"

	"github.com/bits-and-blooms/bitset"
	"golang.org/x/exp/slog"
)

// Minimize
// Minimizes (and determinizes if not already deterministic) the given automaton using Hopcroft's algorithm.
func Minimize(a *Automaton, determinizeWorkLimit int) (*Automaton, error) {
	if a.GetNumStates() == 0 || (a.IsAccept(0) == false && a.GetNumTransitionsWithState(0) == 0) {
		// Fastmatch for common case
		return NewAutomaton(), nil
	}

	var err error
	a, err = determinize(a, determinizeWorkLimit)
	if err != nil {
		return nil, err
	}
	if a.GetNumTransitionsWithState(0) == 1 {
		t := NewTransition()
		a.getTransition(0, 0, t)
		if t.Dest == 0 && t.Min == 0 && t.Max == unicode.MaxRune {
			// Accepts all strings
			return a, nil
		}
	}
	a, err = totalize(a)
	if err != nil {
		return nil, err
	}

	// initialize data structures
	sigma := a.GetStartPoints()
	sigmaLen := len(sigma)
	statesLen := a.GetNumStates()

	reverse := make([][][]int, statesLen)
	for i := range reverse {
		reverse[i] = make([][]int, sigmaLen)
	}

	partition := make([]map[int]struct{}, statesLen)
	for i := range partition {
		m := make(map[int]struct{})
		partition[i] = m
	}

	splitblock := make([][]int, statesLen)
	for i := range splitblock {
		arr := make([]int, 0)
		splitblock[i] = arr
	}

	// block: []int
	block := make([]int, statesLen)

	// active: [][]*StateList
	active := make([][]*StateList, statesLen)
	for i := range active {
		active[i] = make([]*StateList, sigmaLen)
		for j := range active[i] {
			active[i][j] = &StateList{}
		}
	}

	// active2: [][]*StateListNode
	active2 := make([][]*StateListNode, statesLen)
	for i := range active2 {
		active2[i] = make([]*StateListNode, sigmaLen)
	}

	pending := make([]*IntPair, 0)
	pending2 := bitset.New(uint(sigmaLen * statesLen))
	split := bitset.New(uint(statesLen))
	refine := bitset.New(uint(statesLen))
	refine2 := bitset.New(uint(statesLen))

	for q := 0; q < statesLen; q++ {
		splitblock[q] = make([]int, 0)
		partition[q] = make(map[int]struct{})
		for x := 0; x < sigmaLen; x++ {
			active[q][x] = NewStateList()
		}
	}

	// find initial partition and reverse edges
	transition := NewTransition()
	for q := 0; q < statesLen; q++ {
		j := 1
		if a.IsAccept(q) {
			j = 0
		}
		partition[j][q] = struct{}{}
		block[q] = j
		transition.Source = q
		transition.TransitionUpto = -1
		for x := 0; x < sigmaLen; x++ {
			r := reverse[a.Next(transition, sigma[x])]
			if r[x] == nil {
				r[x] = make([]int, 0)
			}
			r[x] = append(r[x], q)
		}
	}

	// initialize active sets
	for j := 0; j <= 1; j++ {
		for x := 0; x < sigmaLen; x++ {
			for q := range partition[j] {
				if reverse[q][x] != nil {
					active2[q][x] = active[j][x].Add(q)
				}
			}
		}
	}

	// initialize pending
	for x := 0; x < sigmaLen; x++ {
		j := 1
		if active[0][x].size <= active[1][x].size {
			j = 0
		}

		pending = append(pending, NewIntPair(j, x))
		pending2.Set(uint(x*statesLen + j))
	}

	// process pending until fixed point
	k := 2
	for len(pending) != 0 {
		//System.out.println("  cycle pending");
		ip := pending[0]
		pending = pending[1:]

		p := ip.n1
		x := ip.n2
		//System.out.println("    pop n1=" + ip.n1 + " n2=" + ip.n2);
		pending2.Clear(uint(x*statesLen + p))

		// find states that need to be split off their blocks
		for m := active[p][x].first; m != nil; m = m.next {
			r := reverse[m.q][x]
			if r != nil {
				for _, i := range r {
					ui := uint(i)
					if !split.Test(ui) {
						split.Set(ui)
						j := block[i]
						splitblock[j] = append(splitblock[j], i)
						uj := uint(j)
						if !refine2.Test(uj) {
							refine2.Set(uj)
							refine.Set(uj)
						}
					}
				}
			}
		}

		j := 0
		// refine blocks
		for ; ; j++ {
			nextJ, ok := refine.NextSet(uint(j))
			if !ok {
				break
			}
			j = int(nextJ)

			sb := splitblock[j]
			if len(sb) < len(partition[j]) {
				b1 := partition[j]
				b2 := partition[k]
				for _, s := range sb {
					delete(b1, s)
					b2[s] = struct{}{}
					block[s] = k
					for c := 0; c < sigmaLen; c++ {
						sn := active2[s][c]
						if sn != nil && sn.sl == active[j][c] {
							sn.Remove()
							active2[s][c] = active[k][c].Add(s)
						}
					}
				}
				// update pending
				for c := 0; c < sigmaLen; c++ {
					aj := active[j][c].size
					ak := active[k][c].size
					ofs := c * statesLen
					if !pending2.Test(uint(ofs+j)) && 0 < aj && aj <= ak {
						pending2.Set(uint(ofs + j))
						pending = append(pending, NewIntPair(j, c))
					} else {
						pending2.Set(uint(ofs + k))
						pending = append(pending, NewIntPair(k, c))
					}
				}
				k++
			}
			refine2.Clear(uint(j))
			for _, s := range sb {
				split.Clear(uint(s))
			}
			// sb.clear();
			splitblock[j] = splitblock[j][:0]
		}
		refine.ClearAll()
	}

	result := NewAutomaton()

	t := NewTransition()

	// make a new state for each equivalence class, set initial state
	stateMap := make([]int, statesLen)
	stateRep := make([]int, k)

	result.CreateState()

	for n := 0; n < k; n++ {

		isInitial := false
		for q := range partition[n] {
			if q == 0 {
				isInitial = true
				//System.out.println("    isInitial!");
				break
			}
		}

		var newState int
		if isInitial {
			newState = 0
		} else {
			newState = result.CreateState()
		}

		for q := range partition[n] {
			stateMap[q] = newState
			result.SetAccept(newState, a.IsAccept(q))
			stateRep[newState] = q // select representative
		}
	}

	// build transitions and set acceptance
	for n := 0; n < k; n++ {
		numTransitions := a.InitTransition(stateRep[n], t)
		for i := 0; i < numTransitions; i++ {
			a.GetNextTransition(t)
			//System.out.println("  add trans");
			if err := result.AddTransition(n, stateMap[t.Dest], t.Min, t.Max); err != nil {
				return nil, err
			}
		}
	}
	result.FinishState()
	slog.Debug("after FinishState minimize", "states", result.GetNumStates())

	return removeDeadStates(result)
}

type IntPair struct {
	n1 int
	n2 int
}

func NewIntPair(n1 int, n2 int) *IntPair {
	return &IntPair{n1: n1, n2: n2}
}

type StateList struct {
	size        int
	first, last *StateListNode
}

func NewStateList() *StateList {
	return &StateList{}
}

func (sl *StateList) Add(q int) *StateListNode {
	return NewStateListNode(q, sl)
}

type StateListNode struct {
	sl         *StateList
	q          int
	next, prev *StateListNode
}

func (n *StateListNode) Remove() {
	n.sl.size--
	if n.sl.first == n {
		n.sl.first = n.next
	} else {
		n.prev.next = n.next
	}
	if n.sl.last == n {
		n.sl.last = n.prev
	} else {
		n.next.prev = n.prev
	}
}

func NewStateListNode(q int, sl *StateList) *StateListNode {
	node := &StateListNode{q: q, sl: sl}
	if sl.size == 0 {
		sl.first = node
	} else {
		sl.last.next = node
		node.prev = sl.last
	}
	sl.size++
	sl.last = node
	return node
}
