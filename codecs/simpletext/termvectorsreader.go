package simpletext

import (
	"bytes"
	"context"
	index2 "github.com/geange/lucene-go/core/interface/index"
	"io"
	"strconv"

	"github.com/geange/gods-generic/maps/treemap"
	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.TermVectorsReader = &TermVectorsReader{}

// TermVectorsReader Reads plain-text term vectors.
// FOR RECREATIONAL USE ONLY
// lucene.experimental
type TermVectorsReader struct {
	offsets []int64
	in      store.IndexInput
	scratch *bytes.Buffer
}

func NewTermVectorsReader(ctx context.Context, directory store.Directory, si *index.SegmentInfo, ioContext *store.IOContext) (*TermVectorsReader, error) {

	fileName := store.SegmentFileName(si.Name(), "", VECTORS_EXTENSION)
	in, err := directory.OpenInput(ctx, fileName)
	if err != nil {
		return nil, err
	}
	reader := &TermVectorsReader{in: in}

	maxDoc, err := si.MaxDoc()
	if err != nil {
		return nil, err
	}
	if err := reader.readIndex(maxDoc); err != nil {
		return nil, err
	}
	return reader, nil
}

func (s *TermVectorsReader) readIndex(size int) error {
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
	return utils.CheckFooter(input)
}

func (s *TermVectorsReader) readLine() error {
	s.scratch.Reset()
	return utils.ReadLine(s.in, s.scratch)
}

func (s *TermVectorsReader) Close() error {
	if err := s.in.Close(); err != nil {
		return err
	}
	s.in = nil
	s.offsets = nil
	return nil
}

func (s *TermVectorsReader) Get(ctx context.Context, doc int) (index2.Fields, error) {
	fields := treemap.New[string, index2.Terms]()
	if _, err := s.in.Seek(s.offsets[doc], io.SeekStart); err != nil {
		return nil, err
	}

	r := utils.NewTextReader(s.in, s.scratch)

	value, err := r.ReadLabel(VECTORS_NUMFIELDS)
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
		_, err := r.ReadLabel(VECTORS_FIELD)
		if err != nil {
			return nil, err
		}

		fieldName, err := r.ReadLabel(VECTORS_FIELDNAME)
		if err != nil {
			return nil, err
		}

		value, err = r.ReadLabel(VECTORS_FIELDPOSITIONS)
		if err != nil {
			return nil, err
		}
		positions, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}

		value, err = r.ReadLabel(VECTORS_FIELDOFFSETS)
		if err != nil {
			return nil, err
		}
		offsets, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}

		value, err = r.ReadLabel(VECTORS_FIELDPAYLOADS)
		if err != nil {
			return nil, err
		}
		payloads, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}

		value, err = r.ReadLabel(VECTORS_FIELDTERMCOUNT)
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
			value, err := r.ReadLabel(VECTORS_TERMTEXT)
			if err != nil {
				return nil, err
			}
			postings := NewSimpleTVPostings()

			terms.terms.Put([]byte(value), postings)
			value, err = r.ReadLabel(VECTORS_TERMFREQ)
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
						v, err := r.ReadLabel(VECTORS_POSITION)
						if err != nil {
							return nil, err
						}
						postings.positions[k], err = strconv.Atoi(v)
						if err != nil {
							return nil, err
						}
						if payloads {
							value, err := r.ReadLabel(VECTORS_PAYLOAD)
							if err != nil {
								return nil, err
							}

							if len(v) != 0 {
								postings.payloads[k] = []byte(value)
							}
						}
					}

					if offsets {
						value, err := r.ReadLabel(VECTORS_STARTOFFSET)
						if err != nil {
							return nil, err
						}
						postings.positions[k], err = strconv.Atoi(value)
						if err != nil {
							return nil, err
						}

						value, err = r.ReadLabel(VECTORS_ENDOFFSET)
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

func (s *TermVectorsReader) CheckIntegrity() error {
	return nil
}

func (s *TermVectorsReader) Clone(context.Context) index.TermVectorsReader {
	return &TermVectorsReader{
		offsets: s.offsets,
		in:      s.in.Clone().(store.IndexInput),
		scratch: bytes.NewBuffer(s.scratch.Bytes()),
	}
}

func (s *TermVectorsReader) GetMergeInstance() index.TermVectorsReader {
	return s
}
