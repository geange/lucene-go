package search

import (
	"fmt"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/interface/search"
	"github.com/geange/lucene-go/core/util/structure"
)

// BaseTopDocs
// Represents hits returned by IndexSearcher.search(Query, int).
type BaseTopDocs struct {
	// The total number of hits for the query.
	totalHits *search.TotalHits

	// The top hits for the query.
	scoreDocs []search.ScoreDoc
}

func (t *BaseTopDocs) GetTotalHits() *search.TotalHits {
	return t.totalHits
}

func (t *BaseTopDocs) GetScoreDocs() []search.ScoreDoc {
	return t.scoreDocs
}

// NewTopDocs Constructs a TopDocs.
func NewTopDocs(totalHits *search.TotalHits, scoreDocs []search.ScoreDoc) *BaseTopDocs {
	return &BaseTopDocs{totalHits: totalHits, scoreDocs: scoreDocs}
}

func MergeTopDocs(start, topN int, shardHits []search.TopDocs, setShardIndex bool) (search.TopDocs, error) {
	return mergeAuxTopDocs(nil, start, topN, shardHits, setShardIndex)
}

// Auxiliary method used by the merge impls.
// A sort value of null is used to indicate that docs should be sorted by score.
func mergeAuxTopDocs(sort index.Sort, start, size int, shardHits []search.TopDocs, setShardIndex bool) (search.TopDocs, error) {
	if sort == nil {
		queue := NewScoreMergeSortQueue(shardHits)
		totalHitCount := int64(0)
		totalHitsRelation := search.EQUAL_TO
		availHitCount := 0
		for shardIDX := 0; shardIDX < len(shardHits); shardIDX++ {
			shard := shardHits[shardIDX]
			// totalHits can be non-zero even if no hits were
			// collected, when searchAfter was used:
			totalHitCount += shard.GetTotalHits().Value
			// If any hit count is a lower bound then the merged
			// total hit count is a lower bound as well
			if shard.GetTotalHits().Relation == search.GREATER_THAN_OR_EQUAL_TO {
				totalHitsRelation = search.GREATER_THAN_OR_EQUAL_TO
			}
			if shard.GetScoreDocs() != nil && len(shard.GetScoreDocs()) > 0 {
				availHitCount += len(shard.GetScoreDocs())
				queue.Add(NewShardRef(shardIDX, setShardIndex == false))
			}
		}

		var hits []search.ScoreDoc
		if availHitCount <= start {
			hits = make([]search.ScoreDoc, 0)
		} else {
			hits = make([]search.ScoreDoc, min(size, availHitCount-start))
			requestedResultWindow := start + size
			numIterOnHits := min(availHitCount, requestedResultWindow)
			hitUpto := 0

			for hitUpto < numIterOnHits {
				//assert queue.size() > 0;
				ref := queue.Top()
				hit := shardHits[ref.shardIndex].GetScoreDocs()[ref.hitIndex]
				ref.hitIndex++

				if setShardIndex {
					// caller asked us to record shardIndex (index of the TopDocs array) this hit is coming from:
					hit.SetShardIndex(ref.shardIndex)
				} else if hit.GetShardIndex() == -1 {
					return nil, fmt.Errorf("setShardIndex is false but TopDocs[%d].scoreDocs[%d] is not set", ref.shardIndex, ref.hitIndex-1)
				}

				if hitUpto >= start {
					hits[hitUpto-start] = hit
				}

				hitUpto++

				if ref.hitIndex < len(shardHits[ref.shardIndex].GetScoreDocs()) {
					// Not done with this these TopDocs yet:
					queue.UpdateTop()
				} else {
					queue.Pop()
				}
			}
		}

		totalHits := search.NewTotalHits(totalHitCount, totalHitsRelation)
		if sort == nil {
			return NewTopDocs(totalHits, hits), nil
		}
		return NewTopFieldDocs(totalHits, hits, sort.GetSort()), nil
	} else {
		panic("")
	}
}

// ShardRef
// Refers to one hit:
type ShardRef struct {
	// Which shard (index into shardHits[]):
	shardIndex int

	// True if we should use the incoming ScoreDoc.shardIndex for sort order
	useScoreDocIndex bool

	// Which hit within the shard:
	hitIndex int
}

func NewShardRef(shardIndex int, useScoreDocIndex bool) *ShardRef {
	return &ShardRef{
		shardIndex:       shardIndex,
		useScoreDocIndex: useScoreDocIndex,
	}
}

func (s *ShardRef) GetShardIndex(scoreDoc search.ScoreDoc) int {
	if s.useScoreDocIndex {
		if scoreDoc.GetShardIndex() == -1 {
			//throw new IllegalArgumentException("setShardIndex is false but TopDocs[" + shardIndex + "].scoreDocs[" + hitIndex + "] is not set");
		}
		return scoreDoc.GetShardIndex()
	} else {
		// NOTE: we don't assert that shardIndex is -1 here, because caller could in fact have set it but asked us to ignore it now
		return s.shardIndex
	}
}

type ScoreMergeSortQueue struct {
	*structure.PriorityQueue[*ShardRef]

	shardHits [][]search.ScoreDoc
}

func NewScoreMergeSortQueue(shardHits []search.TopDocs) *ScoreMergeSortQueue {
	queue := &ScoreMergeSortQueue{
		shardHits: make([][]search.ScoreDoc, len(shardHits)),
	}
	for i := range queue.shardHits {
		queue.shardHits[i] = shardHits[i].GetScoreDocs()
	}

	queue.PriorityQueue = structure.NewPriorityQueue(len(shardHits), func(first, second *ShardRef) bool {
		//assert first != second;
		firstScoreDoc := queue.shardHits[first.shardIndex][first.hitIndex]
		secondScoreDoc := queue.shardHits[second.shardIndex][second.hitIndex]
		if firstScoreDoc.GetScore() < secondScoreDoc.GetScore() {
			return false
		} else if firstScoreDoc.GetScore() > secondScoreDoc.GetScore() {
			return true
		} else {
			return tieBreakLessThan(first, firstScoreDoc, second, secondScoreDoc)
		}
	})
	return queue
}

func tieBreakLessThan(first *ShardRef, firstDoc search.ScoreDoc, second *ShardRef, secondDoc search.ScoreDoc) bool {
	firstShardIndex := first.GetShardIndex(firstDoc)
	secondShardIndex := second.GetShardIndex(secondDoc)
	// Tie break: earlier shard wins
	if firstShardIndex < secondShardIndex {
		return true
	} else if firstShardIndex > secondShardIndex {
		return false
	} else {
		// Tie break in same shard: resolve however the
		// shard had resolved it:
		//assert first.hitIndex != second.hitIndex
		return first.hitIndex < second.hitIndex
	}
}
