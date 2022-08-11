package index

import "github.com/geange/lucene-go/core/store"

// MultiLevelSkipListWriter This abstract class writes skip lists with multiple levels.
//
//	 Example for skipInterval = 3:
//	                                                    c            (skip level 2)
//	                c                 c                 c            (skip level 1)
//	    x     x     x     x     x     x     x     x     x     x      (skip level 0)
//	d d d d d d d d d d d d d d d d d d d d d d d d d d d d d d d d  (posting list)
//	    3     6     9     12    15    18    21    24    27    30     (df)
//
//	d - document
//	x - skip data
//	c - skip data with child pointer
//
//	Skip level i contains every skipInterval-th entry from skip level i-1.
//	Therefore the number of entries on level i is: floor(df / ((skipInterval ^ (i + 1))).
//
//	Each skip entry on a level  i>0 contains a pointer to the corresponding skip entry in list i-1.
//	This guarantees a logarithmic amount of skips to find the target document.
//
//	While this class takes care of writing the different skip levels,
//	subclasses must define the actual format of the skip data.
type MultiLevelSkipListWriter interface {
	MultiLevelSkipListWriterBase
	MultiLevelSkipListWriterExt
}

type MultiLevelSkipListWriterBase interface {
	// WriteSkipData Subclasses must implement the actual skip data encoding in this method.
	// Params: 	level – the level skip data shall be writing for
	//			skipBuffer – the skip buffer to write to
	WriteSkipData(level int, skipBuffer store.IndexOutput) error
}

type MultiLevelSkipListWriterExt interface {
	// Init Allocates internal skip buffers.
	Init()

	// ResetSkip Creates new buffers or empties the existing ones
	ResetSkip() error

	// BufferSkip Writes the current skip data to the buffers. The current document frequency
	// determines the max level is skip data is to be written to.
	// Params: 	df – the current document frequency
	// Throws: 	IOException – If an I/O error occurs
	BufferSkip(df int) error

	// WriteSkip Writes the buffered skip lists to the given output.
	// Params: 	output – the IndexOutput the skip lists shall be written to
	// Returns: the pointer the skip list starts
	WriteSkip(output store.IndexOutput) error

	// WriteLevelLength Writes the length of a level to the given output.
	// Params: 	levelLength – the length of a level
	//			output – the IndexOutput the length shall be written to
	WriteLevelLength(levelLength int64, output store.IndexOutput) error

	// WriteChildPointer Writes the child pointer of a block to the given output.
	// Params: 	childPointer – block of higher level point to the lower level
	//			skipBuffer – the skip buffer to write to
	WriteChildPointer(childPointer int64, skipBuffer store.DataOutput) error
}

//var _ MultiLevelSkipListWriterExt = &MultiLevelSkipListWriterImp{}

type MultiLevelSkipListWriterImp struct {
	// number of levels in this skip list
	numberOfSkipLevels int

	// the skip interval in the list with level = 0
	skipInterval int

	// skipInterval used for level > 0
	skipMultiplier int

	// for every skip level a different buffer is used
	skipBuffer []store.RAMOutputStream
}

func (m *MultiLevelSkipListWriterImp) WriteSkip(output store.IndexOutput) error {
	//TODO implement me
	panic("implement me")
}

func (m *MultiLevelSkipListWriterImp) BufferSkip(df int) error {
	//TODO implement me
	panic("implement me")
}
