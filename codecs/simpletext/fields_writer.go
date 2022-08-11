package simpletext

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"strconv"
)

var (
	FieldsWriterToken = struct {
		END          []byte
		FIELD        []byte
		TERM         []byte
		DOC          []byte
		FREQ         []byte
		POS          []byte
		START_OFFSET []byte
		END_OFFSET   []byte
		PAYLOAD      []byte
	}{
		END:          []byte("END"),
		FIELD:        []byte("field "),
		TERM:         []byte("  term "),
		DOC:          []byte("    doc "),
		FREQ:         []byte("      freq "),
		POS:          []byte("      pos "),
		START_OFFSET: []byte("      startOffset "),
		END_OFFSET:   []byte("      endOffset "),
		PAYLOAD:      []byte("        payload "),
	}
)

var _ index.FieldsConsumer = &FieldsWriter{}

type FieldsWriter struct {
	*index.FieldsConsumerImp

	out        store.IndexOutput
	writeState *index.SegmentWriteState
	segment    string
	docCount   int

	skipWriter                   *SimpleTextSkipWriter
	competitiveImpactAccumulator *index.CompetitiveImpactAccumulator
	lastDocFilePointer           int64
}

func (s *FieldsWriter) Close() error {
	//TODO implement me
	panic("implement me")
}

func (s *FieldsWriter) Write(fields index.Fields, norms index.NormsProducer) error {
	//TODO implement me
	panic("implement me")
}

func (s *FieldsWriter) WriteV1(fieldInfos index.FieldInfos, fields index.Fields,
	normsProducer index.NormsProducer) error {

	for _, field := range fields.Names() {
		terms, err := fields.Terms(field)
		if err != nil {
			return err
		}
		if terms == nil {
			continue
		}

		fieldInfo := fieldInfos.FieldInfo(field)

		wroteField := false

		hasPositions := terms.HasPositions()
		hasFreqs := terms.HasFreqs()
		hasPayloads := fieldInfo.HasPayloads()
		hasOffsets := terms.HasOffsets()
		fieldHasNorms := fieldInfo.HasNorms()

		var norms index.NumericDocValues
		if fieldHasNorms && normsProducer != nil {
			norms, err = normsProducer.GetNorms(fieldInfo)
			if err != nil {
				return err
			}
		}

		flags := 0
		if hasPositions {
			flags = index.POSTINGS_ENUM_POSITIONS
			if hasPayloads {
				flags = flags | index.POSTINGS_ENUM_PAYLOADS
			}
			if hasOffsets {
				flags = flags | index.POSTINGS_ENUM_OFFSETS
			}
		} else {
			if hasFreqs {
				flags = flags | index.POSTINGS_ENUM_FREQS
			}
		}

		termsEnum, err := terms.Iterator()
		if err != nil {
			return err
		}
		var postingsEnum index.PostingsEnum

		// for each term in field
		for {
			term, err := termsEnum.Next()
			if err != nil {
				return err
			}

			if term == nil {
				break
			}

			docCount := 0
			if err := s.skipWriter.ResetSkip(); err != nil {
				return err
			}
			s.competitiveImpactAccumulator.Clear()
			s.lastDocFilePointer = -1

			postingsEnum, err = termsEnum.Postings(postingsEnum, flags)
			if err != nil {
				return err
			}

			wroteTerm := false

			// for each doc in field+term
			for {
				doc, err := postingsEnum.NextDoc()
				if err != nil {
					break
				}
				if doc == index.NO_MORE_DOCS {
					break
				}

				if !wroteTerm {

					if !wroteField {
						// we lazily do this, in case the field had no terms
						s.write(FieldsWriterToken.FIELD)
						s.write([]byte(field))
						s.newline()
						wroteField = true
					}

					// we lazily do this, in case the term had
					// zero docs
					s.write(FieldsWriterToken.TERM)
					s.write(term)
					s.newline()
					wroteTerm = true
				}
				if s.lastDocFilePointer == -1 {
					s.lastDocFilePointer = s.out.GetFilePointer()
				}
				s.write(FieldsWriterToken.DOC)
				s.write([]byte(strconv.Itoa(doc)))
				s.newline()
				if hasFreqs {
					freq, err := postingsEnum.Freq()
					if err != nil {
						return err
					}
					s.write(FieldsWriterToken.FREQ)
					s.write([]byte(strconv.Itoa(freq)))
					s.newline()

					if hasPositions {
						// for assert:
						//lastStartOffset := 0

						// for each pos in field+term+doc
						for i := 0; i < freq; i++ {
							position, err := postingsEnum.NextPosition()
							if err != nil {
								return err
							}

							s.write(FieldsWriterToken.POS)
							s.write([]byte(strconv.Itoa(position)))
							s.newline()

							if hasOffsets {
								startOffset, err := postingsEnum.StartOffset()
								if err != nil {
									return err
								}
								endOffset, err := postingsEnum.EndOffset()
								if err != nil {
									return err
								}

								//lastStartOffset = startOffset
								s.write(FieldsWriterToken.START_OFFSET)
								s.write([]byte(strconv.Itoa(startOffset)))
								s.newline()
								s.write(FieldsWriterToken.END_OFFSET)
								s.write([]byte(strconv.Itoa(endOffset)))
								s.newline()
							}

							payload, err := postingsEnum.GetPayload()

							if payload != nil && len(payload) > 0 {
								s.write(FieldsWriterToken.PAYLOAD)
								s.write(payload)
								s.newline()
							}
						}
					}
					norm, err := s.getNorm(doc, norms)
					if err != nil {
						return err
					}
					s.competitiveImpactAccumulator.Add(freq, norm)
				} else {
					norm, err := s.getNorm(doc, norms)
					if err != nil {
						return err
					}
					s.competitiveImpactAccumulator.Add(1, norm)
				}
				docCount++
				if docCount != 0 && docCount%BLOCK_SIZE == 0 {
					s.skipWriter.bufferSkip(doc, s.lastDocFilePointer, docCount, s.competitiveImpactAccumulator)
					s.competitiveImpactAccumulator.Clear()
					s.lastDocFilePointer = -1
				}
			}
			if docCount >= BLOCK_SIZE {
				if err := s.skipWriter.WriteSkip(s.out); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *FieldsWriter) getNorm(doc int, norms index.NumericDocValues) (int64, error) {
	if norms == nil {
		return 1, nil
	}
	found, err := norms.AdvanceExact(doc)
	if err != nil {
		return 0, err
	}
	if !found {
		return 1, nil
	}
	return norms.LongValue()
}

func (s *FieldsWriter) write(field []byte) error {
	return s.out.WriteBytes(field)
}

func (s *FieldsWriter) newline() error {
	return s.out.WriteByte('\n')
}
