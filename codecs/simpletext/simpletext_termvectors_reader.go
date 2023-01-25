package simpletext

import (
	"bytes"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"io"
	"strconv"
)

var _ index.TermVectorsReader = &SimpleTextTermVectorsReader{}

// SimpleTextTermVectorsReader Reads plain-text term vectors.
// FOR RECREATIONAL USE ONLY
// lucene.experimental
type SimpleTextTermVectorsReader struct {
	offsets []int64
	in      store.IndexInput
	scratch *bytes.Buffer
}

func NewSimpleTextTermVectorsReader(directory store.Directory, si *index.SegmentInfo,
	context *store.IOContext) (*SimpleTextTermVectorsReader, error) {

	fileName := store.SegmentFileName(si.Name(), "", VECTORS_EXTENSION)
	in, err := directory.OpenInput(fileName, context)
	if err != nil {
		return nil, err
	}
	reader := &SimpleTextTermVectorsReader{in: in}

	maxDoc, err := si.MaxDoc()
	if err != nil {
		return nil, err
	}
	if err := reader.readIndex(maxDoc); err != nil {
		return nil, err
	}
	return reader, nil
}

func (s *SimpleTextTermVectorsReader) readIndex(size int) error {
	input := store.NewBufferedChecksumIndexInput(s.in)
	s.offsets = make([]int64, 0, size)
	upto := 0
	for !bytes.Equal(s.scratch.Bytes(), VECTORS_END) {
		if err := s.readLine(); err != nil {
			return err
		}

		if bytes.HasPrefix(s.scratch.Bytes(), VECTORS_DOC) {
			s.offsets = append(s.offsets, input.GetFilePointer())
			upto++
		}

	}
	return CheckFooter(input)
}

func (s *SimpleTextTermVectorsReader) readLine() error {
	s.scratch.Reset()
	return ReadLine(s.in, s.scratch)
}

func (s *SimpleTextTermVectorsReader) Close() error {
	if err := s.in.Close(); err != nil {
		return err
	}
	s.in = nil
	s.offsets = nil
	return nil
}

func (s *SimpleTextTermVectorsReader) Get(doc int) (index.Fields, error) {
	fields := treemap.NewWithStringComparator()
	if _, err := s.in.Seek(s.offsets[doc], io.SeekStart); err != nil {
		return nil, err
	}

	value, err := readValue(s.in, VECTORS_NUMFIELDS, s.scratch)
	if err != nil {
		return nil, err
	}
	numFields, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	if numFields == 0 {
		return nil, nil
	}

	for i := 0; i < numFields; i++ {
		// skip fieldNumber:
		_, err := readValue(s.in, VECTORS_FIELD, s.scratch)
		if err != nil {
			return nil, err
		}

		fieldName, err := readValue(s.in, VECTORS_FIELDNAME, s.scratch)
		if err != nil {
			return nil, err
		}

		value, err = readValue(s.in, VECTORS_FIELDPOSITIONS, s.scratch)
		if err != nil {
			return nil, err
		}
		positions, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}

		value, err = readValue(s.in, VECTORS_FIELDOFFSETS, s.scratch)
		if err != nil {
			return nil, err
		}
		offsets, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}

		value, err = readValue(s.in, VECTORS_FIELDPAYLOADS, s.scratch)
		if err != nil {
			return nil, err
		}
		payloads, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}

		value, err = readValue(s.in, VECTORS_FIELDTERMCOUNT, s.scratch)
		if err != nil {
			return nil, err
		}
		termCount, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}

		terms := NewSimpleTVTerms(offsets, positions, payloads)
		fields.Put(fieldName, terms)

		for j := 0; j < termCount; j++ {
			value, err := readValue(s.in, VECTORS_TERMTEXT, s.scratch)
			if err != nil {
				return nil, err
			}
			postings := NewSimpleTVPostings()

			terms.terms.Put(value, postings)
			value, err = readValue(s.in, VECTORS_TERMFREQ, s.scratch)
			if err != nil {
				return nil, err
			}
			freq, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			postings.freq = freq

			if positions || offsets {
				if positions {
					postings.positions = make([]int, postings.freq)
					if payloads {
						postings.payloads = make([][]byte, postings.freq)
					}
				}

				if offsets {
					postings.startOffsets = make([]int, postings.freq)
					postings.endOffsets = make([]int, postings.freq)
				}

				for k := 0; k < postings.freq; k++ {
					if positions {
						v, err := readValue(s.in, VECTORS_POSITION, s.scratch)
						if err != nil {
							return nil, err
						}
						postings.positions[k], err = strconv.Atoi(v)
						if err != nil {
							return nil, err
						}
						if payloads {
							value, err := readValue(s.in, VECTORS_PAYLOAD, s.scratch)
							if err != nil {
								return nil, err
							}

							if len(v) != 0 {
								postings.payloads[k] = []byte(value)
							}
						}
					}

					if offsets {
						value, err := readValue(s.in, VECTORS_STARTOFFSET, s.scratch)
						if err != nil {
							return nil, err
						}
						postings.positions[k], err = strconv.Atoi(value)
						if err != nil {
							return nil, err
						}

						value, err = readValue(s.in, VECTORS_ENDOFFSET, s.scratch)
						if err != nil {
							return nil, err
						}
						postings.endOffsets[k], err = strconv.Atoi(value)
						if err != nil {
							return nil, err
						}
					}
				}
			}
		}
	}
	return NewSimpleTVFields(fields), nil
}

func (s *SimpleTextTermVectorsReader) CheckIntegrity() error {
	return nil
}

func (s *SimpleTextTermVectorsReader) Clone() index.TermVectorsReader {
	return &SimpleTextTermVectorsReader{
		offsets: s.offsets,
		in:      s.in.Clone(),
		scratch: bytes.NewBuffer(s.scratch.Bytes()),
	}
}

func (s *SimpleTextTermVectorsReader) GetMergeInstance() index.TermVectorsReader {
	return s
}
