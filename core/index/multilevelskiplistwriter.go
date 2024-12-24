package index

import (
	"context"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
)

// MultiLevelSkipListWriter
// This abstract class writes skip lists with multiple levels.
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
	// WriteSkipData
	// Subclasses must implement the actual skip data encoding in this method.
	// Params: 	level – the level skip data shall be writing for
	//			skipBuffer – the skip buffer to write to
	WriteSkipData(level int, skipBuffer store.IndexOutput) error

	// Init Allocates internal skip buffers.
	Init()

	// ResetSkip
	// Creates new buffers or empties the existing ones
	ResetSkip()

	// BufferSkip
	// Writes the current skip data to the buffers. The current document frequency
	// determines the max level is skip data is to be written to.
	// Params: 	df – the current document frequency
	// Throws: 	IOException – If an I/O error occurs
	BufferSkip(df int) error

	// WriteSkip
	// Writes the buffered skip lists to the given output.
	// Params: 	output – the IndexOutput the skip lists shall be written to
	// Returns: the pointer the skip list starts
	WriteSkip(output store.IndexOutput) (int64, error)

	// WriteLevelLength
	// Writes the length of a level to the given output.
	// Params: 	levelLength – the length of a level
	//			output – the IndexOutput the length shall be written to
	WriteLevelLength(levelLength int64, output store.IndexOutput) error

	// WriteChildPointer
	// Writes the child pointer of a block to the given output.
	// Params: 	childPointer – block of higher level point to the lower level
	//			skipBuffer – the skip buffer to write to
	WriteChildPointer(childPointer int64, skipBuffer store.DataOutput) error
}

type MultiLevelSkipListWriterContext struct {
	NumberOfSkipLevels int
	SkipInterval       int
	SkipMultiplier     int
	SkipBuffer         []*store.BufferOutput
}

func NewMultiLevelSkipListWriterContext(skipInterval, skipMultiplier, maxSkipLevels, df int) *MultiLevelSkipListWriterContext {
	var numberOfSkipLevels int
	if df <= skipInterval {
		numberOfSkipLevels = 1
	} else {
		numberOfSkipLevels = 1 + util.Log(df/skipInterval, skipMultiplier)
	}

	if numberOfSkipLevels > maxSkipLevels {
		numberOfSkipLevels = maxSkipLevels
	}

	return &MultiLevelSkipListWriterContext{
		SkipInterval:       skipInterval,
		NumberOfSkipLevels: numberOfSkipLevels,
		SkipMultiplier:     skipMultiplier,
	}
}

func (m *MultiLevelSkipListWriterContext) init() {
	m.SkipBuffer = m.SkipBuffer[:0]
	for i := 0; i < m.NumberOfSkipLevels; i++ {
		m.SkipBuffer = append(m.SkipBuffer, store.NewBufferDataOutput())
	}
}

func (m *MultiLevelSkipListWriterContext) ResetSkip() {
	if len(m.SkipBuffer) == 0 {
		m.init()
		return
	}

	for _, output := range m.SkipBuffer {
		output.Reset()
	}
}

func (m *MultiLevelSkipListWriterContext) BufferSkip(ctx context.Context, df int,
	spi MultiLevelSkipListWriterSPI) error {

	numLevels := 1
	df /= m.SkipInterval

	// determine max level
	for (df%m.SkipMultiplier) == 0 && numLevels < m.NumberOfSkipLevels {
		numLevels++
		df /= m.SkipMultiplier
	}

	childPointer := int64(0)

	for level := 0; level < numLevels; level++ {
		if err := spi.WriteSkipData(ctx, level, m.SkipBuffer[level], m); err != nil {
			return err
		}

		newChildPointer := m.SkipBuffer[level].GetFilePointer()

		if level != 0 {
			// store child pointers for all levels except the lowest
			if err := spi.WriteChildPointer(ctx, childPointer, m.SkipBuffer[level]); err != nil {
				return err
			}
		}

		//remember the childPointer for the next level
		childPointer = newChildPointer
	}

	return nil
}

type MultiLevelSkipListWriterSPI interface {
	ResetSkip(mwc *MultiLevelSkipListWriterContext) error
	// WriteSkipData
	// Subclasses must implement the actual skip data encoding in this method.
	// level: the level skip data shall be writing for
	// skipBuffer: the skip buffer to write to
	WriteSkipData(ctx context.Context, level int, skipBuffer store.IndexOutput, mwc *MultiLevelSkipListWriterContext) error

	// WriteSkip
	// Writes the buffered skip lists to the given output.
	// output: the IndexOutput the skip lists shall be written to
	// Returns: the pointer the skip list starts
	WriteSkip(ctx context.Context, output store.IndexOutput, mwc *MultiLevelSkipListWriterContext) (int64, error)

	// WriteLevelLength
	// Writes the length of a level to the given output.
	// levelLength: the length of a level
	// output: the IndexOutput the length shall be written to
	WriteLevelLength(ctx context.Context, levelLength int64, output store.IndexOutput) error

	// WriteChildPointer
	// Writes the child pointer of a block to the given output.
	// childPointer: block of higher level point to the lower level
	// skipBuffer: the skip buffer to write to
	WriteChildPointer(ctx context.Context, childPointer int64, skipBuffer store.DataOutput) error
}

//var _ MultiLevelSkipListWriterExt = &MultiLevelSkipListWriterImp{}

type BaseMultiLevelSkipListWriterConfig struct {
	SkipInterval      int
	SkipMultiplier    int
	MaxSkipLevels     int
	DF                int
	WriteSkipData     func(level int, skipBuffer store.IndexOutput) error
	WriteLevelLength  func(levelLength int64, output store.IndexOutput) error
	WriteChildPointer func(childPointer int64, skipBuffer store.DataOutput) error
}

type BaseMultiLevelSkipListWriter struct {
	// number of levels in this skip list
	numberOfSkipLevels int

	// the skip interval in the list with level = 0
	skipInterval int

	// skipInterval used for level > 0
	skipMultiplier int

	// for every skip level a different buffer is used
	// TODO: fix
	//skipBuffer []*store.RAMOutputStream

	// Subclasses must implement the actual skip data encoding in this method.
	// level: the level skip data shall be writing for
	// skipBuffer: the skip buffer to write to
	writeSkipData func(level int, skipBuffer store.IndexOutput) error

	fnWriteLevelLength  func(levelLength int64, output store.IndexOutput) error
	fnWriteChildPointer func(childPointer int64, skipBuffer store.DataOutput) error
}

func NewBaseMultiLevelSkipListWriter(cfg *BaseMultiLevelSkipListWriterConfig) *BaseMultiLevelSkipListWriter {
	this := &BaseMultiLevelSkipListWriter{}

	this.skipInterval = cfg.SkipInterval
	this.skipMultiplier = cfg.SkipMultiplier

	numberOfSkipLevels := 0
	// calculate the maximum number of skip levels for this document frequency
	if cfg.DF <= cfg.SkipInterval {
		numberOfSkipLevels = 1
	} else {
		numberOfSkipLevels = 1 + util.Log(cfg.DF/cfg.SkipInterval, cfg.SkipMultiplier)
	}

	// make sure it does not exceed maxSkipLevels
	if numberOfSkipLevels > cfg.MaxSkipLevels {
		numberOfSkipLevels = cfg.MaxSkipLevels
	}
	this.SetNumberOfSkipLevels(numberOfSkipLevels)
	return this
}

func (m *BaseMultiLevelSkipListWriter) NumberOfSkipLevels() int {
	return m.numberOfSkipLevels
}

func (m *BaseMultiLevelSkipListWriter) SetNumberOfSkipLevels(numberOfSkipLevels int) {
	m.numberOfSkipLevels = numberOfSkipLevels
}

/*
func (m *MultiLevelSkipListWriterDefault) Init() {
	m.skipBuffer = make([]*store.RAMOutputStream, 0, m.NumberOfSkipLevels)
	for i := 0; i < m.NumberOfSkipLevels; i++ {
		m.skipBuffer = append(m.skipBuffer, store.NewRAMOutputStream())
	}
}

// ResetSkip Creates new buffers or empties the existing ones
func (m *MultiLevelSkipListWriterDefault) ResetSkip() {
	if len(m.skipBuffer) == 0 {
		m.Init()
	} else {
		for i := 0; i < len(m.skipBuffer); i++ {
			m.skipBuffer[i].Reset()
		}
	}
}

// WriteSkip Creates new buffers or empties the existing ones
func (m *MultiLevelSkipListWriterDefault) WriteSkip(output store.IndexOutput) (int64, error) {
	skipPointer := output.GetFilePointer()
	if len(m.skipBuffer) == 0 {
		return skipPointer, nil
	}

	for level := m.NumberOfSkipLevels - 1; level > 0; level-- {
		length := m.skipBuffer[level].GetFilePointer()
		if length > 0 {
			if err := m.writeLevelLength(length, output); err != nil {
				return 0, err
			}
			if err := m.skipBuffer[level].WriteTo(output); err != nil {
				return 0, err
			}
		}
	}
	if err := m.skipBuffer[0].WriteTo(output); err != nil {
		return 0, err
	}

	return skipPointer, nil
}

func (m *MultiLevelSkipListWriterDefault) BufferSkip(df int) error {
	//assert df % skipInterval == 0;
	numLevels := 1
	df /= m.skipInterval

	// determine max level
	for (df%m.skipMultiplier) == 0 && numLevels < m.NumberOfSkipLevels {
		numLevels++
		df /= m.skipMultiplier
	}

	childPointer := 0

	for level := 0; level < numLevels; level++ {
		if err := m.writeSkipData(level, m.skipBuffer[level]); err != nil {
			return err
		}

		newChildPointer := m.skipBuffer[level].GetFilePointer()

		if level != 0 {
			// store child pointers for all levels except the lowest
			if err := m.writeChildPointer(int64(childPointer), m.skipBuffer[level]); err != nil {
				return err
			}
		}

		//remember the childPointer for the next level
		childPointer = int(newChildPointer)
	}
	return nil
}

// Writes the length of a level to the given output.
// levelLength – the length of a level
// output – the IndexOutput the length shall be written to
func (m *MultiLevelSkipListWriterDefault) writeLevelLength(levelLength int64, output store.IndexOutput) error {
	if m.fnWriteLevelLength != nil {
		return m.fnWriteLevelLength(levelLength, output)
	}
	return output.WriteUvarint(nil, uint64(levelLength))
}

// Writes the child pointer of a block to the given output.
// childPointer – block of higher level point to the lower level
// skipBuffer – the skip buffer to write to
func (m *MultiLevelSkipListWriterDefault) writeChildPointer(childPointer int64, skipBuffer store.DataOutput) error {
	if m.fnWriteChildPointer != nil {
		return m.fnWriteChildPointer(childPointer, skipBuffer)
	}
	return skipBuffer.WriteUvarint(nil, uint64(childPointer))
}

*/
