package analysis

import (
	"github.com/geange/lucene-go/core/util/attribute"
)

// FilteringTokenFilter Abstract base class for TokenFilters that may remove tokens. You have to implement
// accept and return a boolean if the current token should be preserved. incrementToken uses this method to
// decide if a token should be passed to the caller.
type FilteringTokenFilter interface {
	TokenFilter

	Acceptable
}

type Acceptable interface {
	Accept() (bool, error)
}

type BaseFilteringTokenFilter struct {
	*BaseTokenFilter

	acceptable       Acceptable
	posIncrAtt       attribute.PositionIncrAttr
	skippedPositions int
}

func NewFilteringTokenFilter(acceptable Acceptable, in TokenStream) *BaseFilteringTokenFilter {
	return &BaseFilteringTokenFilter{
		acceptable:       acceptable,
		BaseTokenFilter:  NewBaseTokenFilter(in),
		posIncrAtt:       in.AttributeSource().PositionIncrement(),
		skippedPositions: 0,
	}
}

func (r *BaseFilteringTokenFilter) IncrementToken() (bool, error) {

	r.skippedPositions = 0
	for {
		ok, err := r.input.IncrementToken()
		if err != nil {
			return false, err
		}

		if !ok {
			break
		}

		ok, err = r.acceptable.Accept()
		if err != nil {
			return false, err
		}
		if ok {
			if r.skippedPositions != 0 {
				err := r.posIncrAtt.SetPositionIncrement(r.posIncrAtt.GetPositionIncrement() + r.skippedPositions)
				if err != nil {
					return false, err
				}
			}
			return true, nil
		}
		r.skippedPositions += r.posIncrAtt.GetPositionIncrement()
	}
	return false, nil
}

func (r *BaseFilteringTokenFilter) Reset() error {
	err := r.input.Reset()
	if err != nil {
		return err
	}
	r.skippedPositions = 0
	return nil
}

func (r *BaseFilteringTokenFilter) End() error {
	err := r.input.End()
	if err != nil {
		return err
	}
	return r.posIncrAtt.SetPositionIncrement(r.posIncrAtt.GetPositionIncrement() + r.skippedPositions)
}
