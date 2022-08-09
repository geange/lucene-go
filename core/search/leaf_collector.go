package search

import (
	"github.com/geange/lucene-go/core/index"
)

// LeafCollector Collector decouples the Score from the collected doc: the Score computation is skipped entirely if it's not needed. Collectors that do need the Score should implement the setScorer method, to hold onto the passed Scorer instance, and call Scorer.Score() within the collect method to compute the current hit's Score. If your collector may request the Score for a single hit multiple times, you should use ScoreCachingWrappingScorer.
// NOTE: The doc that is passed to the collect method is relative to the current reader. If your collector needs to resolve this to the docID space of the Multi*Reader, you must re-base it by recording the docBase from the most recent setNextReader call. Here's a simple example showing how to collect docIDs into a BitSet:
//
//	IndexSearcher searcher = new IndexSearcher(indexReader);
//	final BitSet bits = new BitSet(indexReader.maxDoc());
//	searcher.search(query, new Collector() {
//
//	  public LeafCollector getLeafCollector(LeafReaderContext context)
//	      throws IOException {
//	    final int docBase = context.docBase;
//	    return new LeafCollector() {
//
//	      // ignore scorer
//	      public void setScorer(Scorer scorer) throws IOException {
//	      }
//
//	      public void collect(int doc) throws IOException {
//	        bits.set(docBase + doc);
//	      }
//
//	    };
//	  }
//
//	});
//
// Not all collectors will need to rebase the docID. For example, a collector that simply counts the total
// number of hits would skip it.
type LeafCollector interface {
	// SetScorer Called before successive calls to collect(int). Implementations that need the Score of
	// the current document (passed-in to collect(int)), should save the passed-in Scorer and call
	// scorer.Score() when needed.
	SetScorer(scorer Scorable) error

	// Collect Called once for every document matching a query, with the unbased document number.
	// Note: The collection of the current segment can be terminated by throwing a CollectionTerminatedException.
	// In this case, the last docs of the current org.apache.lucene.index.LeafReaderContext will be skipped
	// and IndexSearcher will swallow the exception and continue collection with the next leaf.
	// Note: This is called in an inner search loop. For good search performance, implementations of this
	// method should not call IndexSearcher.doc(int) or org.apache.lucene.index.IndexReader.document(int) on
	// every hit. Doing so can slow searches by an order of magnitude or more.
	Collect(doc int) error

	// CompetitiveIterator Optionally returns an iterator over competitive documents. Collectors should
	// delegate this method to their comparators if their comparators provide the skipping functionality
	// over non-competitive docs. The default is to return null which is interpreted as the collector
	// provide any competitive iterator.
	CompetitiveIterator() (index.DocIdSetIterator, error)
}

type LeafCollectorImp struct {
}

func (*LeafCollectorImp) CompetitiveIterator() (index.DocIdSetIterator, error) {
	return nil, nil
}
