package simpletext

import (
	"context"
	"errors"
	"io"
	"strconv"

	"github.com/geange/lucene-go/codecs/utils"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
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

var _ index.FieldsConsumer = &TextFieldsWriter{}

type TextFieldsWriter struct {
	out        store.IndexOutput
	writeState *index.SegmentWriteState
	segment    string
	docCount   int

	skipWriter                   *SkipWriter
	competitiveImpactAccumulator *coreIndex.CompetitiveImpactAccumulator
	lastDocFilePointer           int64
}

func NewFieldsWriter(ctx context.Context, writeState *index.SegmentWriteState) (*TextFieldsWriter, error) {
	fileName := getPostingsFileName(writeState.SegmentInfo.Name(), writeState.SegmentSuffix)
	out, err := writeState.Directory.CreateOutput(ctx, fileName)
	if err != nil {
		return nil, err
	}

	sw, err := NewSkipWriter(writeState)
	if err != nil {
		return nil, err
	}
	return &TextFieldsWriter{
		out:                          out,
		writeState:                   writeState,
		segment:                      writeState.SegmentInfo.Name(),
		docCount:                     0,
		skipWriter:                   sw,
		competitiveImpactAccumulator: coreIndex.NewCompetitiveImpactAccumulator(),
		lastDocFilePointer:           0,
	}, nil
}

func (s *TextFieldsWriter) Close() error {
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

func (s *TextFieldsWriter) Write(ctx context.Context, fields index.Fields, norms index.NormsProducer) error {
	return s.writeFields(ctx, s.writeState.FieldInfos, fields, norms)
}

func (s *TextFieldsWriter) writeFields(ctx context.Context, fieldInfos index.FieldInfos, fields index.Fields,
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
			flags = coreIndex.POSTINGS_ENUM_POSITIONS
			if hasPayloads {
				flags = flags | coreIndex.POSTINGS_ENUM_PAYLOADS
			}
			if hasOffsets {
				flags = flags | coreIndex.POSTINGS_ENUM_OFFSETS
			}
		} else {
			if hasFreqs {
				flags = flags | coreIndex.POSTINGS_ENUM_FREQS
			}
		}

		termsEnum, err := terms.Iterator()
		if err != nil {
			return err
		}
		var postingsEnum index.PostingsEnum

		// for each term in field
		for {
			term, err := termsEnum.Next(ctx)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}

			docCount := 0

			err = s.skipWriter.ResetSkip()
			if err != nil {
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
				doc, err := postingsEnum.NextDoc(ctx)
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
					err := s.skipWriter.BufferSkip(ctx, doc, s.lastDocFilePointer, docCount, s.competitiveImpactAccumulator)
					if err != nil {
						return err
					}
					s.competitiveImpactAccumulator.Clear()
					s.lastDocFilePointer = -1
				}
			}
			if docCount >= BLOCK_SIZE {
				if _, err := s.skipWriter.WriteSkip(ctx, s.out); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *TextFieldsWriter) getNorm(doc int, norms index.NumericDocValues) (int64, error) {
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

func (s *TextFieldsWriter) writeValue(label, value []byte) error {
	if err := utils.WriteBytes(s.out, label); err != nil {
		return err
	}
	if err := utils.WriteBytes(s.out, value); err != nil {
		return err
	}
	return utils.NewLine(s.out)
}

func (s *TextFieldsWriter) write(field []byte) error {
	return utils.WriteBytes(s.out, field)
}

func (s *TextFieldsWriter) newline() error {
	return utils.NewLine(s.out)
}
