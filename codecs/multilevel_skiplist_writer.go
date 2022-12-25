package codecs

import "github.com/geange/lucene-go/core/store"

// MultiLevelSkipListWriter
/**
This abstract class writes skip lists with multiple levels.

  Example for skipInterval = 3:
                                                      c            (skip level 2)
                  c                 c                 c            (skip level 1)
      x     x     x     x     x     x     x     x     x     x      (skip level 0)
  d d d d d d d d d d d d d d d d d d d d d d d d d d d d d d d d  (posting list)
      3     6     9     12    15    18    21    24    27    30     (df)

  d - document
  x - skip data
  c - skip data with child pointer

  Skip level i contains every skipInterval-th entry from skip level i-1.
  Therefore the number of entries on level i is: floor(df / ((skipInterval ^ (i + 1))).

  Each skip entry on a level i>0 contains a pointer to the corresponding skip entry in list i-1.
  This guarantees a logarithmic amount of skips to find the target document.

  While this class takes care of writing the different skip levels,
  subclasses must define the actual format of the skip data.

  lucene.experimental

*/
type MultiLevelSkipListWriter struct {

	// number of levels in this skip list
	numberOfSkipLevels int

	// the skip interval in the list with level = 0
	skipInterval int

	// skipInterval used for level > 0
	skipMultiplier int

	// for every skip level a different buffer is used
	skipBuffer []store.RAMOutputStream

	// must implement the actual skip data encoding in this method.
	// * level: the level skip data shall be writing for
	// * skipBuffer: the skip buffer to write to
	WriteSkipData func(level int, skipBuffer store.IndexOutput) error
}

func (r *MultiLevelSkipListWriter) Init() {
	r.skipBuffer = make([]store.RAMOutputStream, 0)
	for i := 0; i < r.numberOfSkipLevels; i++ {

		r.skipBuffer = append(r.skipBuffer)
	}
}
