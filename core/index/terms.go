package index

import (
	"context"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/util/automaton"
	"github.com/geange/lucene-go/core/util/bytesref"
)

type TermsSPI interface {
	Iterator() (index.TermsEnum, error)
	Size() (int, error)
}

type BaseTerms struct {
	spi      TermsSPI
	Iterator func() (index.TermsEnum, error)
	Size     func() (int, error)
}

func NewTerms(spi TermsSPI) *BaseTerms {
	return &BaseTerms{spi: spi}
}

func (t *BaseTerms) Intersect(compiled *automaton.CompiledAutomaton, startTerm []byte) (index.TermsEnum, error) {
	// TODO: could we factor out a common interface b/w
	// CompiledAutomaton and FST?  Then we could pass FST there too,
	// and likely speed up resolving terms to deleted docs ... but
	// AutomatonTermsEnum makes this tricky because of its on-the-fly
	// cycle detection

	// TODO: eventually we could support seekCeil/Exact on
	// the returned enum, instead of only being able to seek
	// at the start

	//termsEnum, err := t.DVFUIterator()
	//if err != nil {
	//	return nil, err
	//}
	//
	//if compiled.Type() != automaton.AUTOMATON_TYPE_NORMAL {
	//	return nil, errors.New("please use CompiledAutomaton.getTermsEnum instead")
	//}
	//
	//if len(startTerm) > 0 {
	//	//
	//	//return nAutomatonTermsEnum(termsEnum, compiled);
	//}
	panic("")
}

func (t *BaseTerms) GetMin() ([]byte, error) {
	iterator, err := t.Iterator()
	if err != nil {
		return nil, err
	}
	return iterator.Next(nil)
}

func (t *BaseTerms) GetMax() ([]byte, error) {
	size, err := t.Size()
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return nil, nil
	} else if size >= 0 {
		iterator, err := t.Iterator()
		if err != nil {
			return nil, err
		}
		if err := iterator.SeekExactByOrd(context.TODO(), int64(size-1)); err != nil {
			return nil, err
		}
		return iterator.Term()
	}

	// otherwise: binary search
	iterator, err := t.Iterator()
	if err != nil {
		return nil, err
	}
	v, err := iterator.Next(nil)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}

	scratch := bytesref.NewBytesRefBuilder()
	scratch.AppendByte(0)

	for {
		low := 0
		high := 256

		for low != high {
			mid := (low + high) >> 1
			scratch.SetByteAt(scratch.Length()-1, byte(mid))
			status, err := iterator.SeekCeil(nil, scratch.Get())
			if err != nil {
				return nil, err
			}
			if status == index.SEEK_STATUS_END {
				// Scratch was too high
				if mid == 0 {
					scratch.SetLength(scratch.Length() - 1)
					return scratch.Get(), nil
				}
			} else {
				// Scratch was too low; there is at least one term
				// still after it:
				if low == mid {
					break
				}
				low = mid
			}
		}

		// Recurse to next digit:
		scratch.SetLength(scratch.Length() + 1)
		scratch.Grow(scratch.Length())
	}
}
