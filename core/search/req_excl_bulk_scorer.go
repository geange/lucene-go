package search

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/util"
	"math"
)

var _ BulkScorer = &ReqExclBulkScorer{}

type ReqExclBulkScorer struct {
	req  BulkScorer
	excl index.DocIdSetIterator
}

func newReqExclBulkScorer(req BulkScorer, excl index.DocIdSetIterator) *ReqExclBulkScorer {
	return &ReqExclBulkScorer{
		req:  req,
		excl: excl,
	}
}

func (r *ReqExclBulkScorer) Score(collector LeafCollector, acceptDocs util.Bits) error {
	_, err := r.ScoreRange(collector, acceptDocs, 0, math.MaxInt32)
	return err
}

func (r *ReqExclBulkScorer) ScoreRange(collector LeafCollector, acceptDocs util.Bits, min, max int) (int, error) {
	upTo := min
	exclDoc := r.excl.DocID()

	var err error

	for upTo < max {
		if exclDoc < upTo {
			exclDoc, err = r.excl.Advance(upTo)
			if err != nil {

			}
		}
		if exclDoc == upTo {
			// upTo is excluded so we can consider that we scored up to upTo+1
			upTo += 1
			exclDoc, err = r.excl.NextDoc()
			if err != nil {
				return 0, err
			}
		} else {
			upTo, err = r.req.ScoreRange(collector, acceptDocs, upTo, util.Min(exclDoc, max))
			if err != nil {
				return 0, err
			}
		}
	}

	if upTo == max {
		upTo, err = r.req.ScoreRange(collector, acceptDocs, upTo, upTo)
		if err != nil {
			return 0, err
		}
	}

	return upTo, nil
}

func (r *ReqExclBulkScorer) Cost() int64 {
	return r.req.Cost()
}
