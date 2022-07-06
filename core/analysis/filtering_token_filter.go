package analysis

import "github.com/geange/lucene-go/core/tokenattributes"

// FilteringTokenFilter Abstract base class for TokenFilters that may remove tokens. You have to implement
// accept and return a boolean if the current token should be preserved. incrementToken uses this method to
// decide if a token should be passed to the caller.
type FilteringTokenFilter interface {
	TokenFilter

	FilteringTokenFilterPLG
}

type FilteringTokenFilterPLG interface {
	Accept() (bool, error)
}

type FilteringTokenFilterIMP struct {
	FilteringTokenFilterPLG

	*TokenFilterImp

	posIncrAtt       tokenattributes.PositionIncrementAttribute
	skippedPositions int
}

func (r *FilteringTokenFilterIMP) IncrementToken() (bool, error) {

	r.skippedPositions = 0
	for {
		ok, err := r.input.IncrementToken()
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

func (r *FilteringTokenFilterIMP) Reset() error {
	err := r.input.Reset()
	if err != nil {
		return err
	}
	r.skippedPositions = 0
	return nil
}

func (r *FilteringTokenFilterIMP) End() error {
	err := r.input.End()
	if err != nil {
		return err
	}
	return r.posIncrAtt.SetPositionIncrement(r.posIncrAtt.GetPositionIncrement() + r.skippedPositions)
}
