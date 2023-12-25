package simpletext

import (
	"errors"

	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.TermVectorsWriter = &TermVectorsWriter{}

var (
	VECTORS_EXTENSION = "vec"

	VECTORS_END            = []byte("END")
	VECTORS_DOC            = []byte("doc ")
	VECTORS_NUMFIELDS      = []byte("  numfields ")
	VECTORS_FIELD          = []byte("  field ")
	VECTORS_FIELDNAME      = []byte("    name ")
	VECTORS_FIELDPOSITIONS = []byte("    positions ")
	VECTORS_FIELDOFFSETS   = []byte("    offsets   ")
	VECTORS_FIELDPAYLOADS  = []byte("    payloads  ")
	VECTORS_FIELDTERMCOUNT = []byte("    numterms ")
	VECTORS_TERMTEXT       = []byte("    term ")
	VECTORS_TERMFREQ       = []byte("      freq ")
	VECTORS_POSITION       = []byte("      position ")
	VECTORS_PAYLOAD        = []byte("        payload ")
	VECTORS_STARTOFFSET    = []byte("        startoffset ")
	VECTORS_ENDOFFSET      = []byte("        endoffset ")
)

// TermVectorsWriter Writes plain-text term vectors.
// FOR RECREATIONAL USE ONLY
// lucene.experimental
type TermVectorsWriter struct {
	out            store.IndexOutput
	numDocsWritten int
	offsets        bool
	positions      bool
	payloads       bool
}

func NewTermVectorsWriter(dir store.Directory,
	segment string, context *store.IOContext) (*TermVectorsWriter, error) {

	fileName := store.SegmentFileName(segment, "", VECTORS_EXTENSION)
	out, err := dir.CreateOutput(nil, fileName)
	if err != nil {
		return nil, err
	}

	return &TermVectorsWriter{
		out:            out,
		numDocsWritten: 0,
		offsets:        false,
		positions:      false,
		payloads:       false,
	}, nil
}

func (s *TermVectorsWriter) Close() error {
	return s.out.Close()
}

func (s *TermVectorsWriter) StartDocument(numVectorFields int) error {
	if err := writeValue(s.out, VECTORS_DOC, s.numDocsWritten); err != nil {
		return err
	}

	if err := writeValue(s.out, VECTORS_NUMFIELDS, numVectorFields); err != nil {
		return err
	}

	s.numDocsWritten++
	return nil
}

func (s *TermVectorsWriter) FinishDocument() error {
	return nil
}

func (s *TermVectorsWriter) StartField(info *document.FieldInfo, numTerms int, positions, offsets, payloads bool) error {
	if err := writeValue(s.out, VECTORS_FIELD, info.Number()); err != nil {
		return err
	}
	if err := writeValue(s.out, VECTORS_FIELDNAME, info.Name()); err != nil {
		return err
	}
	if err := writeValue(s.out, VECTORS_FIELDPOSITIONS, positions); err != nil {
		return err
	}
	if err := writeValue(s.out, VECTORS_FIELDOFFSETS, offsets); err != nil {
		return err
	}
	if err := writeValue(s.out, VECTORS_FIELDPAYLOADS, payloads); err != nil {
		return err
	}
	if err := writeValue(s.out, VECTORS_FIELDTERMCOUNT, numTerms); err != nil {
		return err
	}

	s.positions = positions
	s.offsets = offsets
	s.payloads = payloads
	return nil
}

func (s *TermVectorsWriter) FinishField() error {
	return nil
}

func (s *TermVectorsWriter) StartTerm(term []byte, freq int) error {
	if err := writeValue(s.out, VECTORS_TERMTEXT, term); err != nil {
		return err
	}

	if err := writeValue(s.out, VECTORS_TERMFREQ, freq); err != nil {
		return err
	}
	return nil
}

func (s *TermVectorsWriter) FinishTerm() error {
	return nil
}

func (s *TermVectorsWriter) AddPosition(position, startOffset, endOffset int, payload []byte) error {
	if s.positions {
		if err := writeValue(s.out, VECTORS_POSITION, position); err != nil {
			return err
		}

		if s.payloads {
			if err := writeValue(s.out, VECTORS_PAYLOAD, payload); err != nil {
				return err
			}
		}
	}

	if s.offsets {
		if err := writeValue(s.out, VECTORS_STARTOFFSET, startOffset); err != nil {
			return err
		}

		if err := writeValue(s.out, VECTORS_ENDOFFSET, endOffset); err != nil {
			return err
		}
	}

	return nil
}

func (s *TermVectorsWriter) Finish(fis *index.FieldInfos, numDocs int) error {
	if s.numDocsWritten != numDocs {
		return errors.New("mergeVectors produced an invalid result")
	}
	if err := utils.WriteBytes(s.out, VECTORS_END); err != nil {
		return err
	}
	if err := utils.NewLine(s.out); err != nil {
		return err
	}
	return utils.WriteChecksum(s.out)
}
