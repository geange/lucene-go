package search

import (
	"context"
	"math"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
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

func (r *ReqExclBulkScorer) Score(collector index.LeafCollector, acceptDocs util.Bits, minDoc, maxDoc int) (int, error) {
	if minDoc < 0 && maxDoc < 0 {
		minDoc = 0
		maxDoc = math.MaxInt32
	}

	upTo := minDoc
	exclDoc := r.excl.DocID()

	var err error

	for upTo < maxDoc {
		if exclDoc < upTo {
			exclDoc, err = r.excl.Advance(context.Background(), upTo)
			if err != nil {
				return 0, err
			}
		}
		if exclDoc == upTo {
			// upTo is excluded so we can consider that we scored up to upTo+1
			upTo += 1
			exclDoc, err = r.excl.NextDoc(context.Background())
			if err != nil {
				return 0, err
			}
		} else {
			upTo, err = r.req.Score(collector, acceptDocs, upTo, min(exclDoc, maxDoc))
			if err != nil {
				return 0, err
			}
		}
	}

	if upTo == maxDoc {
		upTo, err = r.req.Score(collector, acceptDocs, upTo, upTo)
		if err != nil {
			return 0, err
		}
	}

	return upTo, nil
}

func (r *ReqExclBulkScorer) Cost() int64 {
	return r.req.Cost()
}
