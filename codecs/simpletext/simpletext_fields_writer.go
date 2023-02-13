package simpletext

import (
	"errors"
	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"io"
	"strconv"
)

var (
	FIELDS_END          = []byte("END")
	FIELDS_FIELD        = []byte("field ")
	FIELDS_TERM         = []byte("  term ")
	FIELDS_DOC          = []byte("    doc ")
	FIELDS_FREQ         = []byte("      freq ")
	FIELDS_POS          = []byte("      pos ")
	FIELDS_START_OFFSET = []byte("      startOffset ")
	FIELDS_END_OFFSET   = []byte("      endOffset ")
	FIELDS_PAYLOAD      = []byte("        payload ")
)

var _ index.FieldsConsumer = &SimpleTextFieldsWriter{}

type SimpleTextFieldsWriter struct {
	*index.FieldsConsumerDefault // TODO: fix it

	out        store.IndexOutput
	writeState *index.SegmentWriteState
	segment    string
	docCount   int

	skipWriter                   *SimpleTextSkipWriter
	competitiveImpactAccumulator *index.CompetitiveImpactAccumulator
	lastDocFilePointer           int64
}

func NewSimpleTextFieldsWriter(writeState *index.SegmentWriteState) (*SimpleTextFieldsWriter, error) {
	fileName := getPostingsFileName(writeState.SegmentInfo.Name(), writeState.SegmentSuffix)
	out, err := writeState.Directory.CreateOutput(fileName, writeState.Context)
	if err != nil {
		return nil, err
	}

	skipWriter, err := NewSimpleTextSkipWriter(writeState)
	if err != nil {
		return nil, err
	}
	return &SimpleTextFieldsWriter{
		FieldsConsumerDefault:        nil,
		out:                          out,
		writeState:                   writeState,
		segment:                      writeState.SegmentInfo.Name(),
		docCount:                     0,
		skipWriter:                   skipWriter,
		competitiveImpactAccumulator: index.NewCompetitiveImpactAccumulator(),
		lastDocFilePointer:           0,
	}, nil
}

func (s *SimpleTextFieldsWriter) Close() error {
	if err := s.write(FIELDS_END); err != nil {
		return err
	}
	if err := s.newline(); err != nil {
		return err
	}
	if err := utils.WriteChecksum(s.out); err != nil {
		return err
	}
	return s.out.Close()
}

func (s *SimpleTextFieldsWriter) Write(fields index.Fields, norms index.NormsProducer) error {
	return s.WriteV1(s.writeState.FieldInfos, fields, norms)
}

func (s *SimpleTextFieldsWriter) WriteV1(fieldInfos *index.FieldInfos, fields index.Fields,
	normsProducer index.NormsProducer) error {

	names := fields.Names()

	for _, field := range names {
		terms, err := fields.Terms(field)
		if err != nil {
			if errors.Is(err, io.EOF) {
				continue
			}
			return err
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
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}

			docCount := 0
			s.skipWriter.ResetSkip()
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
					if errors.Is(err, io.EOF) {
						break
					}
					return err
				}

				if !wroteTerm {

					if !wroteField {
						// we lazily do this, in case the field had no terms
						if err := s.writeValue(FIELDS_FIELD, []byte(field)); err != nil {
							return err
						}

						wroteField = true
					}

					// we lazily do this, in case the term had
					// zero docs
					if err := s.writeValue(FIELDS_TERM, term); err != nil {
						return err
					}

					wroteTerm = true
				}
				if s.lastDocFilePointer == -1 {
					s.lastDocFilePointer = s.out.GetFilePointer()
				}
				if err := s.writeValue(FIELDS_DOC, []byte(strconv.Itoa(doc))); err != nil {
					return err
				}

				if hasFreqs {
					freq, err := postingsEnum.Freq()
					if err != nil {
						return err
					}

					if err := s.writeValue(FIELDS_FREQ, []byte(strconv.Itoa(freq))); err != nil {
						return err
					}

					if hasPositions {
						// for assert:
						//lastStartOffset := 0

						// for each pos in field+term+doc
						for i := 0; i < freq; i++ {
							position, err := postingsEnum.NextPosition()
							if err != nil {
								return err
							}

							if err := s.writeValue(FIELDS_POS, []byte(strconv.Itoa(position))); err != nil {
								return err
							}

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
								if err := s.writeValue(FIELDS_START_OFFSET, []byte(strconv.Itoa(startOffset))); err != nil {
									return err
								}

								if err := s.writeValue(FIELDS_END_OFFSET, []byte(strconv.Itoa(endOffset))); err != nil {
									return err
								}
							}

							payload, err := postingsEnum.GetPayload()

							if payload != nil && len(payload) > 0 {
								if err := s.writeValue(FIELDS_PAYLOAD, payload); err != nil {
									return err
								}
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
				if _, err := s.skipWriter.WriteSkip(s.out); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *SimpleTextFieldsWriter) getNorm(doc int, norms index.NumericDocValues) (int64, error) {
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

func (s *SimpleTextFieldsWriter) writeValue(label, value []byte) error {
	if err := utils.WriteBytes(s.out, label); err != nil {
		return err
	}
	if err := utils.WriteBytes(s.out, value); err != nil {
		return err
	}
	return utils.WriteNewline(s.out)
}

func (s *SimpleTextFieldsWriter) write(field []byte) error {
	return utils.WriteBytes(s.out, field)
}

func (s *SimpleTextFieldsWriter) newline() error {
	return utils.WriteNewline(s.out)
}
