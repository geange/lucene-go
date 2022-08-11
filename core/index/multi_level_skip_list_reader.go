package index

import (
	"errors"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	"math"
)

// MultiLevelSkipListReader This abstract class reads skip lists with multiple levels. See
// MultiLevelSkipListWriter for the information about the encoding of the multi level skip lists.
// Subclasses must implement the abstract method readSkipData(int, IndexInput) which defines
// the actual format of the skip data.
type MultiLevelSkipListReader interface {
	// ReadSkipData Subclasses must implement the actual skip data encoding in this method.
	// Params: 	level – the level skip data shall be read from
	//			skipStream – the skip stream to read from
	ReadSkipData(level int, skipStream store.IndexInput) (int, error)
}

type MultiLevelSkipListReaderImp struct {
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

	// Doc id of current skip entry per level.
	skipDoc []int

	// Doc id of last read skip entry with docId <= target.
	lastDoc int

	// Child pointer of current skip entry per level.
	childPointer []int64

	// childPointer of last read skip entry with docId <= target.
	lastChildPointer int64

	// childPointer of last read skip entry with docId <= target.
	inputIsBuffered bool
	skipMultiplier  int
}

func NewMultiLevelSkipListReaderImp(skipStream store.IndexInput, maxSkipLevels, skipInterval, skipMultiplier int) *MultiLevelSkipListReaderImp {
	reader := &MultiLevelSkipListReaderImp{
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
	if _, ok := skipStream.(store.BufferedIndexInput); ok {
		reader.inputIsBuffered = true
	}

	for i := 1; i < maxSkipLevels; i++ {
		reader.skipInterval[i] = reader.skipInterval[i-1] * skipMultiplier
	}
	return reader

}

func (m *MultiLevelSkipListReaderImp) GetDoc() int {
	return m.lastDoc
}

func (m *MultiLevelSkipListReaderImp) SkipTo(target int) (int, error) {
	// walk up the levels until highest level is found that has a skip
	// for this target
	level := 0
	for level < m.numberOfSkipLevels-1 && target > m.skipDoc[level+1] {
		level++
	}

	for level >= 0 {
		if target > m.skipDoc[level] {
			if ok, err := m.loadNextSkip(level); err == nil && !ok {
				continue
			}
		} else {
			// no more skips on this level, go down one level
			if level > 0 && m.lastChildPointer > m.skipStream[level-1].GetFilePointer() {
				if err := m.seekChild(level - 1); err != nil {
					return 0, err
				}
			}
			level--
		}
	}

	return m.numSkipped[0] - m.skipInterval[0] - 1, nil
}

func (m *MultiLevelSkipListReaderImp) loadNextSkip(level int) (bool, error) {
	// we have to skip, the target document is greater than the current
	// skip list entry
	m.setLastSkipData(level)

	m.numSkipped[level] += m.skipInterval[level]

	// numSkipped may overflow a signed int, so compare as unsigned.
	if m.numSkipped[level] > m.docCount {
		// this skip list is exhausted
		m.skipDoc[level] = math.MaxInt32
		if m.numberOfSkipLevels > level {
			m.numberOfSkipLevels = level
		}
		return false, nil
	}

	// read next skip entry
	data, err := m.readSkipData(level, m.skipStream[level])
	if err != nil {
		return false, err
	}
	m.skipDoc[level] += int(data)

	if level != 0 {
		// read the child pointer if we are not on the leaf level
		pointer, err := m.readChildPointer(m.skipStream[level])
		if err != nil {
			return false, err
		}
		m.childPointer[level] = pointer + m.skipPointer[level-1]
	}

	return true, nil
}

func (m *MultiLevelSkipListReaderImp) seekChild(level int) error {
	if err := m.skipStream[level].Seek(m.lastChildPointer); err != nil {
		return err
	}
	m.numSkipped[level] = m.numSkipped[level+1] - m.skipInterval[level+1]
	m.skipDoc[level] = m.lastDoc
	if level > 0 {
		pointer, err := m.readChildPointer(m.skipStream[level])
		if err != nil {
			return err
		}
		m.childPointer[level] = pointer + m.skipPointer[level-1]
	}
	return nil
}

func (m *MultiLevelSkipListReaderImp) Close() error {
	for _, input := range m.skipStream {
		if err := input.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (m *MultiLevelSkipListReaderImp) Init(skipPointer int64, df int) error {
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

	return m.loadSkipLevels()
}

// Loads the skip levels
func (m *MultiLevelSkipListReaderImp) loadSkipLevels() error {
	if m.docCount <= m.skipInterval[0] {
		m.numberOfSkipLevels = 1
	} else {
		m.numberOfSkipLevels = 1 + util.Log(m.docCount/m.skipInterval[0], m.skipMultiplier)
	}

	if m.numberOfSkipLevels > m.maxNumberOfSkipLevels {
		m.numberOfSkipLevels = m.maxNumberOfSkipLevels
	}

	m.skipStream[0].Seek(m.skipPointer[0])

	toBuffer := m.numberOfLevelsToBuffer

	for i := m.numberOfSkipLevels - 1; i > 0; i-- {
		// the length of the current level
		length, err := m.readLevelLength(m.skipStream[0])
		if err != nil {
			return err
		}

		// the start pointer of the current level
		m.skipPointer[i] = m.skipStream[0].GetFilePointer()
		if toBuffer > 0 {
			// buffer this level
			m.skipStream[i], err = NewSkipBuffer(m.skipStream[0], int(length))
			if err != nil {
				return err
			}
			toBuffer--
		} else {
			// clone this stream, it is already at the start of the current level
			m.skipStream[i] = m.skipStream[0].Clone()
			if m.inputIsBuffered && length < store.BUFFER_SIZE {
				m.skipStream[i].(store.BufferedIndexInput).SetBufferSize(util.Max(store.MIN_BUFFER_SIZE, int(length)))
			}

			// move base stream beyond the current level
			m.skipStream[0].Seek(m.skipStream[0].GetFilePointer() + length)
		}
	}

	// use base stream for the lowest level
	m.skipPointer[0] = m.skipStream[0].GetFilePointer()

	return nil
}

// Subclasses must implement the actual skip data encoding in this method.
// Params:
// level – the level skip data shall be read from
// skipStream – the skip stream to read from
func (m *MultiLevelSkipListReaderImp) readSkipData(level int, skipStream store.IndexInput) (int64, error) {
	num, err := skipStream.(store.DataInputExt).ReadUvarint()
	return int64(num), err
}

// read the length of the current level written via MultiLevelSkipListWriter.writeLevelLength(long, IndexOutput).
// Params: skipStream – the IndexInput the length shall be read from
// Returns: level length
func (m *MultiLevelSkipListReaderImp) readLevelLength(skipStream store.IndexInput) (int64, error) {
	num, err := skipStream.(store.DataInputExt).ReadUvarint()
	return int64(num), err
}

// read the child pointer written via MultiLevelSkipListWriter.writeChildPointer(long, DataOutput).
// Params: skipStream – the IndexInput the child pointer shall be read from
// Returns: child pointer
func (m *MultiLevelSkipListReaderImp) readChildPointer(skipStream store.IndexInput) (int64, error) {
	num, err := skipStream.(store.DataInputExt).ReadUvarint()
	return int64(num), err
}

func (m *MultiLevelSkipListReaderImp) setLastSkipData(level int) {
	m.lastDoc = m.skipDoc[level]
	m.lastChildPointer = m.childPointer[level]
}

var _ store.IndexInput = &SkipBuffer{}

// SkipBuffer used to buffer the top skip levels
type SkipBuffer struct {
	data    []byte
	pointer int64
	pos     int
}

func (s *SkipBuffer) Clone() store.IndexInput {
	//TODO implement me
	panic("implement me")
}

func NewSkipBuffer(input store.IndexInput, length int) (*SkipBuffer, error) {
	buffer := &SkipBuffer{
		data:    make([]byte, length),
		pointer: input.GetFilePointer(),
	}
	if err := input.ReadBytes(buffer.data); err != nil {
		return nil, err
	}
	return buffer, nil
}

func (s *SkipBuffer) ReadByte() (byte, error) {
	b := s.data[s.pos]
	s.pos++
	return b, nil
}

func (s *SkipBuffer) ReadBytes(b []byte) error {
	copy(b, s.data[s.pos:])
	s.pos += len(b)
	return nil
}

func (s *SkipBuffer) Close() error {
	s.data = nil
	return nil
}

func (s *SkipBuffer) GetFilePointer() int64 {
	return s.pointer + int64(s.pos)
}

func (s *SkipBuffer) Seek(pos int64) error {
	s.pos = int(pos - s.pointer)
	return nil
}

func (s *SkipBuffer) Length() int64 {
	return int64(len(s.data))
}

func (s *SkipBuffer) Slice(_ string, _, _ int64) (store.IndexInput, error) {
	return nil, errors.New("unsupported")
}
