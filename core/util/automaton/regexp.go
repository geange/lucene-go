package automaton

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type Kind int

const (
	REGEXP_UNION         = Kind(iota) // The union of two expressions
	REGEXP_CONCATENATION              // A sequence of two expressions
	REGEXP_INTERSECTION               // The intersection of two expressions
	REGEXP_OPTIONAL                   // An optional expression
	REGEXP_REPEAT                     // An expression that repeats
	REGEXP_REPEAT_MIN                 // An expression that repeats a minimum number of times
	REGEXP_REPEAT_MINMAX              // An expression that repeats a minimum and maximum number of times
	REGEXP_COMPLEMENT                 // The complement of an expression
	REGEXP_CHAR                       // A Character
	REGEXP_CHAR_RANGE                 // A Character range
	REGEXP_ANYCHAR                    // Any Character allowed
	REGEXP_EMPTY                      // An empty expression
	REGEXP_STRING                     // A string expression
	REGEXP_ANYSTRING                  // Any string allowed
	REGEXP_AUTOMATON                  // An Automaton expression
	REGEXP_INTERVAL                   // An Interval expression
)

const (
	INTERSECTION           = 0x0001
	COMPLEMENT             = 0x0002
	EMPTY                  = 0x0004
	ANYSTRING              = 0x0008
	AUTOMATON              = 0x0010
	INTERVAL               = 0x0020
	ALL                    = 0xff
	NONE                   = 0x0000
	ASCII_CASE_INSENSITIVE = 0x0100
)

type RegExp struct {
	kind             Kind
	exp1, exp2       *RegExp
	s                *string
	c                int
	min, max, digits int
	from, to         int
	originalString   []rune
	flags            int
	pos              int
}

type regExpOption struct {
	syntaxFlags int
	matchFlags  int
}
type RegExpOption func(*regExpOption)

func WithSyntaxFlags(syntaxFlags int) RegExpOption {
	return func(option *regExpOption) {
		option.syntaxFlags = syntaxFlags
	}
}

func WithMatchFlags(matchFlags int) RegExpOption {
	return func(option *regExpOption) {
		option.matchFlags = matchFlags
	}
}

func NewRegExp(s string, options ...RegExpOption) (*RegExp, error) {
	opts := &regExpOption{
		syntaxFlags: ALL,
		matchFlags:  0,
	}
	for _, fn := range options {
		fn(opts)
	}

	exp := &RegExp{
		originalString: []rune(s),
	}

	if opts.syntaxFlags > ALL {
		return nil, errors.New("illegal syntax flag")
	}

	if opts.matchFlags > 0 && opts.matchFlags <= ALL {
		return nil, errors.New("illegal match flag")
	}
	exp.flags = opts.syntaxFlags | opts.matchFlags
	var e *RegExp
	var err error
	if len(s) == 0 {
		e = makeString(exp.flags, "")
	} else {
		e, err = exp.parseUnionExp()
		if err != nil {
			return nil, err
		}
		if exp.pos < len(exp.originalString) {
			return nil, fmt.Errorf("end-of-string expected at position %d", exp.pos)
		}
	}
	exp.kind = e.kind
	exp.exp1 = e.exp1
	exp.exp2 = e.exp2
	exp.s = e.s
	exp.c = e.c
	exp.min = e.min
	exp.max = e.max
	exp.digits = e.digits
	exp.from = e.from
	exp.to = e.to
	return exp, nil
}

func newRegExp(flags int, kind Kind, exp1, exp2 *RegExp, s *string, c, min, max, digits, from, to int) *RegExp {
	return &RegExp{
		kind:           kind,
		exp1:           exp1,
		exp2:           exp2,
		s:              s,
		c:              c,
		min:            min,
		max:            max,
		digits:         digits,
		from:           from,
		to:             to,
		originalString: nil,
		flags:          flags,
		pos:            0,
	}
}

func newContainerNode(flags int, kind Kind, exp1, exp2 *RegExp) *RegExp {
	return newRegExp(flags, kind, exp1, exp2, nil, 0, 0, 0, 0, 0, 0)
}

func newRepeatingNode(flags int, kind Kind, exp *RegExp, min, max int) *RegExp {
	return newRegExp(flags, kind, exp, nil, nil, 0, min, max, 0, 0, 0)
}

func newLeafNode(flags int, kind Kind, s *string, c, min, max, digits, from, to int) *RegExp {
	return newRegExp(flags, kind, nil, nil, s, c, min, max, digits, from, to)
}

func makeUnion(flags int, exp1, exp2 *RegExp) *RegExp {
	return newContainerNode(flags, REGEXP_UNION, exp1, exp2)
}

func makeConcatenation(flags int, exp1, exp2 *RegExp) *RegExp {
	if (exp1.kind == REGEXP_CHAR || exp1.kind == REGEXP_STRING) &&
		(exp2.kind == REGEXP_CHAR || exp2.kind == REGEXP_STRING) {
		return makeStringRegExp(flags, exp1, exp2)
	}

	var rexp1, rexp2 *RegExp
	if exp1.kind == REGEXP_CONCATENATION &&
		(exp1.exp2.kind == REGEXP_CHAR || exp1.exp2.kind == REGEXP_STRING) &&
		(exp2.kind == REGEXP_CHAR || exp2.kind == REGEXP_STRING) {
		rexp1 = exp1.exp1
		rexp2 = makeStringRegExp(flags, exp1.exp2, exp2)

	} else if (exp1.kind == REGEXP_CHAR || exp1.kind == REGEXP_STRING) &&
		exp2.kind == REGEXP_CONCATENATION &&
		(exp2.exp1.kind == REGEXP_CHAR || exp2.exp1.kind == REGEXP_STRING) {
		rexp1 = makeStringRegExp(flags, exp1, exp2.exp1)
		rexp2 = exp2.exp2
	} else {
		rexp1 = exp1
		rexp2 = exp2
	}
	return newContainerNode(flags, REGEXP_CONCATENATION, rexp1, rexp2)
}

func makeStringRegExp(flags int, exp1, exp2 *RegExp) *RegExp {
	b := new(bytes.Buffer)
	if exp1.kind == REGEXP_STRING {
		b.WriteString(*exp1.s)
	} else {
		b.WriteRune(rune(exp1.c))
	}

	if exp2.kind == REGEXP_STRING {
		b.WriteString(*exp2.s)
	} else {
		b.WriteRune(rune(exp2.c))
	}

	return makeString(flags, b.String())
}

func makeIntersection(flags int, exp1, exp2 *RegExp) *RegExp {
	return newContainerNode(flags, REGEXP_INTERSECTION, exp1, exp2)
}

func makeOptional(flags int, exp *RegExp) *RegExp {
	return newContainerNode(flags, REGEXP_OPTIONAL, exp, nil)
}

func makeRepeat(flags int, exp *RegExp) *RegExp {
	return newContainerNode(flags, REGEXP_REPEAT, exp, nil)
}

func makeRepeatMin(flags int, exp *RegExp, min int) *RegExp {
	return newRepeatingNode(flags, REGEXP_REPEAT_MIN, exp, min, 0)
}

func makeRepeatRange(flags int, exp *RegExp, min, max int) *RegExp {
	return newRepeatingNode(flags, REGEXP_REPEAT_MINMAX, exp, min, max)
}

func makeComplement(flags int, exp *RegExp) *RegExp {
	return newContainerNode(flags, REGEXP_COMPLEMENT, exp, nil)
}

func makeChar(flags int, c int) *RegExp {
	return newLeafNode(flags, REGEXP_CHAR, nil, c, 0, 0, 0, 0, 0)
}

func makeCharRange(flags, from, to int) (*RegExp, error) {
	if from > to {
		return nil, errors.New("invalid range")
	}
	return newLeafNode(flags, REGEXP_CHAR_RANGE, nil, 0, 0, 0, 0, from, to), nil
}

func makeAnyChar(flags int) *RegExp {
	return newContainerNode(flags, REGEXP_ANYCHAR, nil, nil)
}

func makeEmpty(flags int) *RegExp {
	return newContainerNode(flags, REGEXP_EMPTY, nil, nil)
}

func makeString(flags int, s string) *RegExp {
	return newLeafNode(flags, REGEXP_STRING, &s, 0, 0, 0, 0, 0, 0)
}

func makeAnyString(flags int) *RegExp {
	return newContainerNode(flags, REGEXP_ANYSTRING, nil, nil)
}

func makeAutomaton(flags int, s string) *RegExp {
	return newLeafNode(flags, REGEXP_AUTOMATON, &s, 0, 0, 0, 0, 0, 0)
}

func makeInterval(flags, min, max, digits int) *RegExp {
	return newLeafNode(flags, REGEXP_INTERVAL, nil, 0, min, max, digits, 0, 0)
}

type Provider func(name string) (*Automaton, error)

type toAutomatonOptions struct {
	automata          map[string]*Automaton
	automatonProvider Provider
}

type ToAutomatonOptions func(*toAutomatonOptions)

func WithAutomata(automata map[string]*Automaton) ToAutomatonOptions {
	return func(options *toAutomatonOptions) {
		options.automata = automata
	}
}

func WithAutomatonProvider(automatonProvider Provider) ToAutomatonOptions {
	return func(options *toAutomatonOptions) {
		options.automatonProvider = automatonProvider
	}
}

func (r *RegExp) ToAutomaton() (*Automaton, error) {
	return r.toAutomaton(DEFAULT_DETERMINIZE_WORK_LIMIT)
}

func (r *RegExp) toAutomaton(determinizeWorkLimit int, options ...ToAutomatonOptions) (*Automaton, error) {
	opts := &toAutomatonOptions{
		automata:          nil,
		automatonProvider: nil,
	}
	for _, fn := range options {
		fn(opts)
	}
	return r.toAutomatonInternal(opts.automata, opts.automatonProvider, determinizeWorkLimit)
}

func (r *RegExp) toAutomatonInternal(automata map[string]*Automaton,
	automatonProvider Provider, determinizeWorkLimit int) (*Automaton, error) {

	list := make([]*Automaton, 0)
	var a *Automaton
	var err error
	switch r.kind {
	case REGEXP_UNION:
		list = make([]*Automaton, 0)
		if err := r.findLeaves(r.exp1, REGEXP_UNION, &list, automata, automatonProvider,
			determinizeWorkLimit); err != nil {
			return nil, err
		}
		if err := r.findLeaves(r.exp2, REGEXP_UNION, &list, automata, automatonProvider,
			determinizeWorkLimit); err != nil {
			return nil, err
		}
		a, err = union(list...)
		if err != nil {
			return nil, err
		}
		a, err = Minimize(a, determinizeWorkLimit)
		if err != nil {
			return nil, err
		}
		break
	case REGEXP_CONCATENATION:
		list = make([]*Automaton, 0)
		err := r.findLeaves(r.exp1, REGEXP_CONCATENATION, &list, automata, automatonProvider, determinizeWorkLimit)
		if err != nil {
			return nil, err
		}
		err = r.findLeaves(r.exp2, REGEXP_CONCATENATION, &list, automata, automatonProvider, determinizeWorkLimit)
		if err != nil {
			return nil, err
		}
		a, err = concatenate(list...)
		if err != nil {
			return nil, err
		}
		a, err = Minimize(a, determinizeWorkLimit)
		if err != nil {
			return nil, err
		}
		break
	case REGEXP_INTERSECTION:
		a1, err := r.exp1.toAutomatonInternal(automata, automatonProvider, determinizeWorkLimit)
		if err != nil {
			return nil, err
		}
		a2, err := r.exp2.toAutomatonInternal(automata, automatonProvider, determinizeWorkLimit)
		if err != nil {
			return nil, err
		}

		a, err = intersection(a1, a2)
		if err != nil {
			return nil, err
		}
		a, err = Minimize(a, determinizeWorkLimit)
		if err != nil {
			return nil, err
		}
		break
	case REGEXP_OPTIONAL:
		a1, err := r.exp1.toAutomatonInternal(automata, automatonProvider, determinizeWorkLimit)
		if err != nil {
			return nil, err
		}

		a, err = optional(a1)
		if err != nil {
			return nil, err
		}
		a, err = Minimize(a, determinizeWorkLimit)
		if err != nil {
			return nil, err
		}
		break
	case REGEXP_REPEAT:
		a1, err := r.exp1.toAutomatonInternal(
			automata, automatonProvider, determinizeWorkLimit)
		if err != nil {
			return nil, err
		}
		a, err = repeat(a1)
		if err != nil {
			return nil, err
		}
		a, err = Minimize(a, determinizeWorkLimit)
		if err != nil {
			return nil, err
		}
		break
	case REGEXP_REPEAT_MIN:
		a, err = r.exp1.toAutomatonInternal(automata, automatonProvider, determinizeWorkLimit)
		if err != nil {
			return nil, err
		}
		minNumStates := (a.GetNumStates() - 1) * r.min
		if minNumStates > determinizeWorkLimit {
			return nil, fmt.Errorf("too complex to determinize: %d", minNumStates)
		}
		a, err = repeatCount(a, r.min)
		if err != nil {
			return nil, err
		}
		a, err = Minimize(a, determinizeWorkLimit)
		if err != nil {
			return nil, err
		}
		break
	case REGEXP_REPEAT_MINMAX:
		a, err = r.exp1.toAutomatonInternal(automata, automatonProvider, determinizeWorkLimit)
		if err != nil {
			return nil, err
		}
		minMaxNumStates := (a.GetNumStates() - 1) * r.max
		if minMaxNumStates > determinizeWorkLimit {
			return nil, fmt.Errorf("too complex to determinize: %d", minMaxNumStates)
		}
		a, err = repeatRange(a, r.min, r.max)
		if err != nil {
			return nil, err
		}

		break
	case REGEXP_COMPLEMENT:
		a1, err := r.exp1.toAutomatonInternal(automata, automatonProvider, determinizeWorkLimit)
		if err != nil {
			return nil, err
		}
		a, err = complement(a1, determinizeWorkLimit)
		if err != nil {
			return nil, err
		}

		a, err = Minimize(a, determinizeWorkLimit)
		if err != nil {
			return nil, err
		}
		break
	case REGEXP_CHAR:
		if r.check(ASCII_CASE_INSENSITIVE) {
			a, err = r.toCaseInsensitiveChar(rune(r.c), determinizeWorkLimit)
			if err != nil {
				return nil, err
			}
		} else {
			a, err = defaultAutomata.MakeChar(int32(r.c))
		}
		break
	case REGEXP_CHAR_RANGE:
		a, err = defaultAutomata.MakeCharRange(int32(r.from), int32(r.to))
		if err != nil {
			return nil, err
		}
		break
	case REGEXP_ANYCHAR:
		a, err = defaultAutomata.MakeAnyChar()
		if err != nil {
			return nil, err
		}
		break
	case REGEXP_EMPTY:
		a = defaultAutomata.MakeEmpty()
		break
	case REGEXP_STRING:
		if r.check(ASCII_CASE_INSENSITIVE) {
			a, err = r.toCaseInsensitiveString(determinizeWorkLimit)
			if err != nil {
				return nil, err
			}
		} else {
			a, err = defaultAutomata.MakeString(*r.s)
			if err != nil {
				return nil, err
			}
		}
		break
	case REGEXP_ANYSTRING:
		a, err = defaultAutomata.MakeAnyString()
		break
	case REGEXP_AUTOMATON:
		var aa *Automaton
		if automata != nil {
			aa = automata[*r.s]
		}
		if aa == nil && automatonProvider != nil {
			aa, err = automatonProvider(*r.s)
			if err != nil {
				return nil, err
			}
		}
		if aa == nil {
			return nil, fmt.Errorf("\"%s\" not found", *r.s)
		}
		a = aa
		break
	case REGEXP_INTERVAL:
		a, err = defaultAutomata.MakeDecimalInterval(r.min, r.max, r.digits)
		break
	}
	return a, nil
}

func (r *RegExp) toCaseInsensitiveChar(codepoint rune, determinizeWorkLimit int) (*Automaton, error) {
	case1, err := defaultAutomata.MakeChar(codepoint)
	if err != nil {
		return nil, err
	}
	// For now we only work with ASCII characters
	if codepoint > 128 {
		return case1, nil
	}
	altCase := codepoint
	if unicode.IsLower(codepoint) {
		altCase = unicode.ToUpper(codepoint)
	}

	var result *Automaton
	if altCase != codepoint {
		case2, err := defaultAutomata.MakeChar(altCase)
		if err != nil {
			return nil, err
		}
		result, err = union(case1, case2)
		if err != nil {
			return nil, err
		}
		result, err = Minimize(result, determinizeWorkLimit)
		if err != nil {
			return nil, err
		}
	} else {
		result = case1
	}
	return result, nil
}

func (r *RegExp) toCaseInsensitiveString(determinizeWorkLimit int) (*Automaton, error) {
	list := make([]*Automaton, 0)

	for _, v := range []rune((*r.s)) {
		a, err := r.toCaseInsensitiveChar(v, determinizeWorkLimit)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}

	automata, err := concatenate(list...)
	if err != nil {
		return nil, err
	}
	return Minimize(automata, determinizeWorkLimit)
}

func (r *RegExp) findLeaves(exp *RegExp, kind Kind, list *[]*Automaton,
	automata map[string]*Automaton, automatonProvider Provider, determinizeWorkLimit int) error {
	if exp.kind == kind {
		if err := r.findLeaves(exp.exp1, kind, list, automata, automatonProvider,
			determinizeWorkLimit); err != nil {
			return err
		}

		if err := r.findLeaves(exp.exp2, kind, list, automata, automatonProvider,
			determinizeWorkLimit); err != nil {
			return err
		}
	} else {
		automaton, err := exp.toAutomatonInternal(automata, automatonProvider,
			determinizeWorkLimit)
		if err != nil {
			return err
		}
		*list = append(*list, automaton)
	}
	return nil
}

func (r *RegExp) more() bool {
	return r.pos < len(r.originalString)
}

func (r *RegExp) peek(s string) bool {
	return r.more() && strings.ContainsRune(s, r.originalString[r.pos])
}

func (r *RegExp) match(c int) bool {
	if r.pos >= len(r.originalString) {
		return false
	}
	if r.originalString[r.pos] == rune(c) {
		r.pos++
		return true
	}
	return false
}

func (r *RegExp) next() (int, error) {
	if !r.more() {
		return 0, io.EOF
	}
	ch := r.originalString[r.pos]
	r.pos++
	return int(ch), nil
}

func (r *RegExp) check(flags int) bool {
	return r.flags&flags != 0
}

func (r *RegExp) parseUnionExp() (*RegExp, error) {
	e, err := r.parseInterExp()
	if err != nil {
		return nil, err
	}
	if r.match('|') {
		e2, err := r.parseUnionExp()
		if err != nil {
			return nil, err
		}
		e = makeUnion(r.flags, e, e2)
	}
	return e, nil
}

func (r *RegExp) parseInterExp() (*RegExp, error) {
	e, err := r.parseConcatExp()
	if err != nil {
		return nil, err
	}
	if r.check(INTERSECTION) && r.match('&') {
		e2, err := r.parseInterExp()
		if err != nil {
			return nil, err
		}
		e = makeIntersection(r.flags, e, e2)
	}
	return e, nil
}

func (r *RegExp) parseConcatExp() (*RegExp, error) {
	e, err := r.parseRepeatExp()
	if err != nil {
		return nil, err
	}
	if r.more() && !r.peek(")|") && (!r.check(INTERSECTION) || !r.peek("&")) {
		e2, err := r.parseConcatExp()
		if err != nil {
			return nil, err
		}
		e = makeConcatenation(r.flags, e, e2)
	}
	return e, nil
}

func (r *RegExp) parseRepeatExp() (*RegExp, error) {
	e, err := r.parseComplExp()
	if err != nil {
		return nil, err
	}

	for r.peek("?*+{") {
		if r.match('?') {
			e = makeOptional(r.flags, e)
		} else if r.match('*') {
			e = makeRepeat(r.flags, e)
		} else if r.match('+') {
			e = makeRepeatMin(r.flags, e, 1)
		} else if r.match('{') {
			start := r.pos
			for r.peek("0123456789") {
				if _, err := r.next(); err != nil {
					return nil, err
				}
			}
			if start == r.pos {
				return nil, fmt.Errorf("integer expected at position %d", r.pos)
			}
			n, err := strconv.Atoi(string(r.originalString[start:r.pos]))
			if err != nil {
				return nil, err
			}
			m := -1
			if r.match(',') {
				start = r.pos
				for r.peek("0123456789") {
					if _, err := r.next(); err != nil {
						return nil, err
					}
				}

				if start != r.pos {
					m, err = strconv.Atoi(string(r.originalString[start:r.pos]))
					if err != nil {
						return nil, err
					}
				} else {
					m = n
				}
			} else {
				m = n
			}

			if !r.match('}') {
				return nil, fmt.Errorf("expected '}' at position %d", r.pos)
			}

			if m == -1 {
				e = makeRepeatMin(r.flags, e, n)
			} else {
				e = makeRepeatRange(r.flags, e, n, m)
			}
		}
	}

	return e, nil
}

func (r *RegExp) parseComplExp() (*RegExp, error) {
	if r.check(COMPLEMENT) && r.match('~') {
		e2, err := r.parseComplExp()
		if err != nil {
			return nil, err
		}
		return makeComplement(r.flags, e2), nil
	}
	return r.parseCharClassExp()
}

func (r *RegExp) parseCharClassExp() (*RegExp, error) {
	if r.match('[') {
		negate := false
		if r.match('^') {
			negate = true
		}
		e, err := r.parseCharClasses()
		if err != nil {
			return nil, err
		}
		if negate {
			e = makeIntersection(r.flags, makeAnyChar(r.flags), makeComplement(r.flags, e))
		}
		if !r.match(']') {
			return nil, fmt.Errorf("expected ']' at position %d", r.pos)
		}
		return e, nil
	}
	return r.parseSimpleExp()
}

func (r *RegExp) parseCharClasses() (*RegExp, error) {
	e, err := r.parseCharClass()
	if err != nil {
		return nil, err
	}
	for r.more() && !r.peek("]") {
		e2, err := r.parseCharClass()
		if err != nil {
			return nil, err
		}
		e = makeUnion(r.flags, e, e2)
	}
	return e, nil
}

func (r *RegExp) parseCharClass() (*RegExp, error) {
	c, err := r.parseCharExp()
	if err != nil {
		return nil, err
	}
	if r.match('-') {
		e2, err := r.parseCharExp()
		if err != nil {
			return nil, err
		}
		return makeCharRange(r.flags, c, e2)
	}
	return makeChar(r.flags, c), err
}

func (r *RegExp) parseSimpleExp() (*RegExp, error) {
	if r.match('.') {
		return makeAnyChar(r.flags), nil
	} else if r.check(EMPTY) && r.match('#') {
		return makeEmpty(r.flags), nil
	} else if r.check(ANYSTRING) && r.match('@') {
		return makeAnyString(r.flags), nil
	} else if r.match('"') {
		//  int start = pos;
		//      while (more() && !peek("\""))
		//        next();
		//      if (!match('"')) throw new IllegalArgumentException(
		//          "expected '\"' at position " + pos);
		//      return makeString(flags, originalString.substring(start, pos - 1));
		start := r.pos
		for r.more() && !r.peek("\"") {
			if _, err := r.next(); err != nil {
				return nil, err
			}
		}
		if !r.match('"') {
			return nil, fmt.Errorf("expected '\\\"' at position %d", r.pos)
		}
		return makeString(r.flags, string(r.originalString[start:r.pos-1])), nil
	} else if r.match('(') {
		if r.match(')') {
			return makeString(r.flags, ""), nil
		}
		e, err := r.parseUnionExp()
		if err != nil {
			return nil, err
		}
		if !r.match(')') {
			return nil, fmt.Errorf("expected ')' at position %d", r.pos)
		}
		return e, nil
	} else if (r.check(AUTOMATON) || r.check(INTERVAL)) && r.match('<') {
		start := r.pos
		for r.more() && !r.peek(">") {
			if _, err := r.next(); err != nil {
				return nil, err
			}
		}

		if !r.match('>') {
			return nil, fmt.Errorf("expected '>' at position %d", r.pos)
		}
		s := string(r.originalString[start : r.pos-1])
		i := strings.IndexRune(s, '-')
		if i == -1 {
			if !r.check(AUTOMATON) {
				return nil, fmt.Errorf("interval syntax error at position %d", r.pos-1)
			}
			return makeAutomaton(r.flags, s), nil
		} else {
			if !r.check(INTERVAL) {
				return nil, fmt.Errorf("illegal identifier at position %d", r.pos-1)
			}

			if i == 0 || i == len(s)-1 || i != strings.LastIndexByte(s, '-') {
				smin := s[:i]
				smax := s[i+1:]
				imin, err := strconv.Atoi(smin)
				if err != nil {
					return nil, err
				}
				imax, err := strconv.Atoi(smax)
				if err != nil {
					return nil, err
				}
				digits := 0
				if len(smin) == len(smax) {
					digits = len(smin)
				}

				if imin > imax {
					imin, imax = imax, imin
				}
				return makeInterval(r.flags, imin, imax, digits), nil
			}
			return nil, fmt.Errorf("interval syntax error at position %d", r.pos-1)
		}
	}

	c, err := r.parseCharExp()
	if err != nil {
		return nil, err
	}
	return makeChar(r.flags, c), nil
}

func (r *RegExp) parseCharExp() (int, error) {
	r.match('\\')
	return r.next()
}
