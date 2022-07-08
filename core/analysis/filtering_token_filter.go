package analysis

import "github.com/geange/lucene-go/core/tokenattributes"

// FilteringTokenFilter Abstract base class for TokenFilters that may remove tokens. You have to implement
// accept and return a boolean if the current token should be preserved. incrementToken uses this method to
// decide if a token should be passed to the caller.
type FilteringTokenFilter interface {
	TokenFilter

	FilteringTokenFilterPlg
}

type FilteringTokenFilterPlg interface {
	Accept() (bool, error)
}

type FilteringTokenFilterImp struct {
	FilteringTokenFilterPlg

	*TokenFilterImp

	posIncrAtt       tokenattributes.PositionIncrementAttribute
	skippedPositions int
}

func NewFilteringTokenFilterImp(plg FilteringTokenFilterPlg, in TokenStream) *FilteringTokenFilterImp {
	return &FilteringTokenFilterImp{
		FilteringTokenFilterPlg: plg,
		TokenFilterImp:          NewTokenFilterImp(in),
		posIncrAtt:              in.AttributeSource().PositionIncrement(),
		skippedPositions:        0,
	}
}

func (r *FilteringTokenFilterImp) IncrementToken() (bool, error) {

	r.skippedPositions = 0
	for {
		ok, err := r.Input.IncrementToken()
		if err != nil {
			return false, err
		}

		if !ok {
			break
		}

		ok, err = r.Accept()
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

func (r *FilteringTokenFilterImp) Reset() error {
	err := r.Input.Reset()
	if err != nil {
		return err
	}
	r.skippedPositions = 0
	return nil
}

func (r *FilteringTokenFilterImp) End() error {
	err := r.Input.End()
	if err != nil {
		return err
	}
	return r.posIncrAtt.SetPositionIncrement(r.posIncrAtt.GetPositionIncrement() + r.skippedPositions)
}
