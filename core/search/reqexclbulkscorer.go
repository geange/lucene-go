package search

import (
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"math"
)

var _ index.BulkScorer = &ReqExclBulkScorer{}

type ReqExclBulkScorer struct {
	req  index.BulkScorer
	excl types.DocIdSetIterator
}

func newReqExclBulkScorer(req index.BulkScorer, excl types.DocIdSetIterator) *ReqExclBulkScorer {
	return &ReqExclBulkScorer{
		req:  req,
		excl: excl,
	}
}

func (r *ReqExclBulkScorer) Score(collector index.LeafCollector, acceptDocs util.Bits) error {
	if _, err := r.ScoreRange(collector, acceptDocs, 0, math.MaxInt32); err != nil {
		return err
	}
	return nil
}

func (r *ReqExclBulkScorer) ScoreRange(collector index.LeafCollector, acceptDocs util.Bits, minDoc, maxDoc int) (int, error) {
	upTo := minDoc
	exclDoc := r.excl.DocID()

	var err error

	for upTo < maxDoc {
		if exclDoc < upTo {
			exclDoc, err = r.excl.Advance(upTo)
			if err != nil {
				return 0, err
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
			upTo, err = r.req.ScoreRange(collector, acceptDocs, upTo, min(exclDoc, maxDoc))
			if err != nil {
				return 0, err
			}
		}
	}

	if upTo == maxDoc {
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
