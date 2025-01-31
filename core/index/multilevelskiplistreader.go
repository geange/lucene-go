package index

import (
	"context"
	"errors"
	"io"
	"math"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
)

// MultiLevelSkipListReader
// This interface reads skip lists with multiple levels. See
// MultiLevelSkipListWriter for the information about the encoding of the multi level skip lists.
// Subclasses must implement the abstract method readSkipData(int, IndexInput) which defines
// the actual format of the skip data.
type MultiLevelSkipListReader interface {
	// ReadSkipData Subclasses must implement the actual skip data encoding in this method.
	// level: the level skip data shall be read from
	// skipStream: the skip stream to read from
	ReadSkipData(level int, skipStream store.IndexInput) (int64, error)
}

type MultiLevelSkipListReaderContext struct {
	//the maximum number of skip levels possible for this index
	maxNumberOfSkipLevels int

	// number of levels in this skip list
	numberOfSkipLevels int

	// Expert: defines the number of top skip levels to buffer in memory.
	// Reducing this number results in less memory usage, but possibly
	// slower performance due to more random I/Os.
	// Please notice that the space each level occupies is limited by
	// the skipInterval. The top level can not contain more than
	// skipLevel entries, the second top level can not contain more
	// than skipLevel^2 entries and so forth.
	numberOfLevelsToBuffer int

	docCount int

	// skipStream for each level.
	skipStream []store.IndexInput

	// The start pointer of each skip level.
	skipPointer []int64

	// skipInterval of each level.
	skipInterval []int

	// Number of docs skipped per level. It's possible for some values to overflow a signed int, but this has been accounted for.
	numSkipped []int

	// doc id of current skip entry per level.
	skipDoc []int

	// doc id of last read skip entry with docId <= target.
	lastDoc int

	// Child pointer of current skip entry per level.
	childPointer []int64

	// childPointer of last read skip entry with docId <= target.
	lastChildPointer int64

	// childPointer of last read skip entry with docId <= target.
	inputIsBuffered bool
	skipMultiplier  int
}

type MultiLevelSkipListReaderSPI interface {
	// ReadSkipData
	// Subclasses must implement the actual skip data encoding in this method.
	ReadSkipData(ctx context.Context, level int, skipStream store.IndexInput, mtx *MultiLevelSkipListReaderContext) (int64, error)

	// ReadLevelLength
	// read the length of the current level written via MultiLevelSkipListWriter. writeLevelLength(long, IndexOutput).
	ReadLevelLength(ctx context.Context, skipStream store.IndexInput, mtx *MultiLevelSkipListReaderContext) (int64, error)

	// ReadChildPointer
	// read the child pointer written via MultiLevelSkipListWriter. writeChildPointer(long, DataOutput).
	ReadChildPointer(ctx context.Context, skipStream store.IndexInput, mtx *MultiLevelSkipListReaderContext) (int64, error)
}

func (m *MultiLevelSkipListReaderContext) Init(ctx context.Context, skipPointer int64, df int, spi MultiLevelSkipListReaderSPI) error {
	m.skipPointer[0] = skipPointer
	m.docCount = df

	for i := range m.skipDoc {
		m.skipDoc[i] = 0
	}

	for i := range m.numSkipped {
		m.numSkipped[i] = 0
	}

	for i := range m.childPointer {
		m.childPointer[i] = 0
	}

	for i := 1; i < m.numberOfSkipLevels; i++ {
		m.skipStream[i] = nil
	}

	return m.loadSkipLevels(ctx, spi)
}

// Loads the skip levels
func (m *MultiLevelSkipListReaderContext) loadSkipLevels(ctx context.Context, spi MultiLevelSkipListReaderSPI) error {
	if m.docCount <= m.skipInterval[0] {
		m.numberOfSkipLevels = 1
	} else {
		m.numberOfSkipLevels = 1 + util.Log(m.docCount/m.skipInterval[0], m.skipMultiplier)
	}

	if m.numberOfSkipLevels > m.maxNumberOfSkipLevels {
		m.numberOfSkipLevels = m.maxNumberOfSkipLevels
	}

	if _, err := m.skipStream[0].Seek(m.skipPointer[0], io.SeekStart); err != nil {
		return err
	}

	toBuffer := m.numberOfLevelsToBuffer

	for i := m.numberOfSkipLevels - 1; i > 0; i-- {
		// the length of the current level
		length, err := spi.ReadLevelLength(ctx, m.skipStream[0], m)
		if err != nil {
			return err
		}

		// the start pointer of the current level
		m.skipPointer[i] = m.skipStream[0].GetFilePointer()
		if toBuffer > 0 {
			// buffer this level
			buffer, err := NewSkipBuffer(m.skipStream[0], int(length))
			if err != nil {
				return err
			}
			m.skipStream[i] = buffer
			toBuffer--
		} else {
			// clone this stream, it is already at the start of the current level
			m.skipStream[i] = m.skipStream[0].Clone().(store.IndexInput)
			//if m.inputIsBuffered && length < store.BUFFER_SIZE {
			//	input, ok := m.skipStream[i].(store.BufferedIndexInput)
			//	if ok {
			//		input.SetBufferSize(max(store.MIN_BUFFER_SIZE, int(length)))
			//	}
			//}

			// move base stream beyond the current level
			if _, err := m.skipStream[0].Seek(m.skipStream[0].GetFilePointer()+length, io.SeekStart); err != nil {
				return err
			}
		}
	}

	// use base stream for the lowest level
	m.skipPointer[0] = m.skipStream[0].GetFilePointer()

	return nil
}

type BaseMultiLevelSkipListReader struct {

	//the maximum number of skip levels possible for this index
	maxNumberOfSkipLevels int

	// number of levels in this skip list
	numberOfSkipLevels int

	// Expert: defines the number of top skip levels to buffer in memory.
	// Reducing this number results in less memory usage, but possibly
	// slower performance due to more random I/Os.
	// Please notice that the space each level occupies is limited by
	// the skipInterval. The top level can not contain more than
	// skipLevel entries, the second top level can not contain more
	// than skipLevel^2 entries and so forth.
	numberOfLevelsToBuffer int

	docCount int

	// skipStream for each level.
	skipStream []store.IndexInput

	// The start pointer of each skip level.
	skipPointer []int64

	// skipInterval of each level.
	skipInterval []int

	// Number of docs skipped per level. It's possible for some values to overflow a signed int, but this has been accounted for.
	numSkipped []int

	// doc id of current skip entry per level.
	skipDoc []int

	// doc id of last read skip entry with docId <= target.
	lastDoc int

	// Child pointer of current skip entry per level.
	childPointer []int64

	// childPointer of last read skip entry with docId <= target.
	lastChildPointer int64

	// childPointer of last read skip entry with docId <= target.
	inputIsBuffered bool
	skipMultiplier  int
}

func NewMultiLevelSkipListReaderContext(skipStream store.IndexInput, maxSkipLevels, skipInterval, skipMultiplier int) *MultiLevelSkipListReaderContext {
	reader := &MultiLevelSkipListReaderContext{
		skipStream:            make([]store.IndexInput, maxSkipLevels),
		skipPointer:           make([]int64, maxSkipLevels),
		childPointer:          make([]int64, maxSkipLevels),
		numSkipped:            make([]int, maxSkipLevels),
		maxNumberOfSkipLevels: maxSkipLevels,
		skipInterval:          make([]int, maxSkipLevels),
		skipMultiplier:        skipMultiplier,
		skipDoc:               make([]int, maxSkipLevels),
	}
	reader.skipStream[0] = skipStream
	reader.skipInterval[0] = skipInterval
	//if _, ok := skipStream.(store.BufferedIndexInput); ok {
	//	reader.inputIsBuffered = true
	//}

	for i := 1; i < maxSkipLevels; i++ {
		reader.skipInterval[i] = reader.skipInterval[i-1] * skipMultiplier
	}
	return reader

}

func (m *MultiLevelSkipListReaderContext) MaxNumberOfSkipLevels() int {
	return m.maxNumberOfSkipLevels
}

func (m *MultiLevelSkipListReaderContext) GetSkipDoc(idx int) int {
	return m.skipDoc[idx]
}

func (m *MultiLevelSkipListReaderContext) GetDoc() int {
	return m.lastDoc
}

func (m *MultiLevelSkipListReaderContext) SkipToWithSPI(ctx context.Context, target int, spi MultiLevelSkipListReaderSPI) (int, error) {
	// walk up the levels until highest level is found that has a skip
	// for this target
	level := 0
	for level < m.numberOfSkipLevels-1 && target > m.skipDoc[level+1] {
		level++
	}

	for level >= 0 {
		if target > m.skipDoc[level] {
			if ok, err := m.loadNextSkip(ctx, level, spi); err == nil && !ok {
				continue
			}
		} else {
			// no more skips on this level, go down one level
			if level > 0 && m.lastChildPointer > m.skipStream[level-1].GetFilePointer() {
				if err := m.seekChild(ctx, level-1, spi); err != nil {
					return 0, err
				}
			}
			level--
		}
	}

	return m.numSkipped[0] - m.skipInterval[0] - 1, nil
}

func (m *MultiLevelSkipListReaderContext) loadNextSkip(ctx context.Context, level int, spi MultiLevelSkipListReaderSPI) (bool, error) {
	// we have to skip, the target document is greater than the current
	// skip list entry
	m.setLastSkipData(level)

	m.numSkipped[level] += m.skipInterval[level]

	// numSkipped may overflow a signed int, so CompareFn as unsigned.
	if m.numSkipped[level] > m.docCount {
		// this skip list is exhausted
		m.skipDoc[level] = math.MaxInt32
		if m.numberOfSkipLevels > level {
			m.numberOfSkipLevels = level
		}
		return false, nil
	}

	// read next skip entry
	data, err := spi.ReadSkipData(ctx, level, m.skipStream[level], m)
	if err != nil {
		return false, err
	}
	m.skipDoc[level] += int(data)

	if level != 0 {
		// read the child pointer if we are not on the leaf level
		pointer, err := spi.ReadChildPointer(ctx, m.skipStream[level], m)
		if err != nil {
			return false, err
		}
		m.childPointer[level] = pointer + m.skipPointer[level-1]
	}

	return true, nil
}

func (m *MultiLevelSkipListReaderContext) seekChild(ctx context.Context, level int, spi MultiLevelSkipListReaderSPI) error {
	if _, err := m.skipStream[level].Seek(m.lastChildPointer, io.SeekStart); err != nil {
		return err
	}
	m.numSkipped[level] = m.numSkipped[level+1] - m.skipInterval[level+1]
	m.skipDoc[level] = m.lastDoc
	if level > 0 {
		pointer, err := spi.ReadChildPointer(ctx, m.skipStream[level], m)
		if err != nil {
			return err
		}
		m.childPointer[level] = pointer + m.skipPointer[level-1]
	}
	return nil
}

func (m *MultiLevelSkipListReaderContext) Close() error {
	for _, input := range m.skipStream {
		if err := input.Close(); err != nil {
			return err
		}
	}
	return nil
}

// ReadSkipData
// Subclasses must implement the actual skip data encoding in this method.
// Params:
// level – the level skip data shall be read from
// skipStream – the skip stream to read from
func (m *BaseMultiLevelSkipListReader) ReadSkipData(level int, skipStream store.IndexInput) (int64, error) {
	num, err := skipStream.ReadUvarint(context.Background())
	return int64(num), err
}

// ReadLevelLength
// read the length of the current level written via MultiLevelSkipListWriter.writeLevelLength(long, IndexOutput).
// Params: skipStream – the IndexInput the length shall be read from
// Returns: level length
func (m *BaseMultiLevelSkipListReader) ReadLevelLength(skipStream store.IndexInput) (int64, error) {
	num, err := skipStream.ReadUvarint(context.Background())
	return int64(num), err
}

// ReadChildPointer
// read the child pointer written via MultiLevelSkipListWriter.writeChildPointer(long, DataOutput).
// Params: skipStream – the IndexInput the child pointer shall be read from
// Returns: child pointer
func (m *BaseMultiLevelSkipListReader) ReadChildPointer(skipStream store.IndexInput) (int64, error) {
	num, err := skipStream.ReadUvarint(context.Background())
	return int64(num), err
}

func (m *MultiLevelSkipListReaderContext) setLastSkipData(level int) {
	m.lastDoc = m.skipDoc[level]
	m.lastChildPointer = m.childPointer[level]
}

var _ store.IndexInput = &SkipBuffer{}

// SkipBuffer used to buffer the top skip levels
type SkipBuffer struct {
	*store.BaseIndexInput

	data    []byte
	pointer int64
	pos     int
}

func (s *SkipBuffer) Clone() store.CloneReader {
	//TODO implement me
	panic("implement me")
}

func NewSkipBuffer(in store.IndexInput, length int) (*SkipBuffer, error) {
	input := &SkipBuffer{
		data:    make([]byte, length),
		pointer: in.GetFilePointer(),
	}
	input.BaseIndexInput = store.NewBaseIndexInput(input)

	if _, err := in.Read(input.data); err != nil {
		return nil, err
	}
	return input, nil
}

func (s *SkipBuffer) ReadByte() (byte, error) {
	b := s.data[s.pos]
	s.pos++
	return b, nil
}

func (s *SkipBuffer) Read(b []byte) (int, error) {
	copy(b, s.data[s.pos:])
	s.pos += len(b)
	return len(b), nil
}

func (s *SkipBuffer) Close() error {
	s.data = nil
	return nil
}

func (s *SkipBuffer) GetFilePointer() int64 {
	return s.pointer + int64(s.pos)
}

func (s *SkipBuffer) Seek(pos int64, whence int) (int64, error) {
	s.pos = int(pos - s.pointer)
	return 0, nil
}

func (s *SkipBuffer) Length() int64 {
	return int64(len(s.data))
}

func (s *SkipBuffer) Slice(sliceDescription string, offset, length int64) (store.IndexInput, error) {
	return nil, errors.New("unsupported")
}
