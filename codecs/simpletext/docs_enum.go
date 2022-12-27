package simpletext

import (
	"bytes"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"strconv"
)

var _ index.ImpactsEnum = &SimpleTextDocsEnum{}

type SimpleTextDocsEnum struct {
	inStart      store.IndexInput
	in           store.IndexInput
	omitTF       bool
	docID        int
	tf           int
	scratch      *bytes.Buffer
	scratchUTF16 *bytes.Buffer
	cost         int
	skipReader   *SimpleTextSkipReader
	nextSkipDoc  int
	seekTo       int64
}

func (r *SimpleTextFieldsReader) NewSimpleTextDocsEnum() *SimpleTextDocsEnum {
	return &SimpleTextDocsEnum{
		inStart:      r.in,
		in:           r.in.Clone(),
		omitTF:       false,
		docID:        -1,
		tf:           0,
		scratch:      nil,
		scratchUTF16: nil,
		cost:         0,
		skipReader:   NewSimpleTextSkipReader(r.in.Clone()),
		nextSkipDoc:  0,
		seekTo:       -1,
	}
}

func (s *SimpleTextDocsEnum) CanReuse(in store.IndexInput) bool {
	return in == s.inStart
}

func (s *SimpleTextDocsEnum) DocID() int {
	return s.docID
}

func (s *SimpleTextDocsEnum) NextDoc() (int, error) {
	return s.Advance(s.docID + 1)
}

func (s *SimpleTextDocsEnum) readDoc() (int, error) {
	if s.docID == index.NO_MORE_DOCS {
		return s.docID, nil
	}
	first := true
	termFreq := 0
	var err error
	for {
		lineStart := s.in.GetFilePointer()
		if err := ReadLine(s.in, s.scratch); err != nil {
			return 0, err
		}

		text := s.scratch.Bytes()

		if bytes.HasPrefix(text, FIELDS_DOC) {
			if !first {
				if _, err := s.in.Seek(lineStart, 0); err != nil {
					return 0, err
				}
				if !s.omitTF {
					s.tf = termFreq
				}
				return s.docID, nil
			}

			s.docID, err = strconv.Atoi(string(text[len(FIELDS_DOC):]))

			termFreq = 0
			first = false
		} else if bytes.HasPrefix(text, FIELDS_FREQ) {
			termFreq, err = strconv.Atoi(string(text[len(FIELDS_FREQ):]))
			if err != nil {
				return 0, err
			}
		} else if bytes.HasPrefix(text, FIELDS_POS) {
			// skip termFreq++;
		} else if bytes.HasPrefix(text, FIELDS_START_OFFSET) {
			// skip
		} else if bytes.HasPrefix(text, FIELDS_END_OFFSET) {
			// skip
		} else if bytes.HasPrefix(text, FIELDS_PAYLOAD) {
			// skip
		} else {

			if !first {
				if _, err := s.in.Seek(lineStart, 0); err != nil {
					return 0, err
				}
				if !s.omitTF {
					s.tf = termFreq
				}
				return s.docID, nil
			}
			s.docID = index.NO_MORE_DOCS
			return s.docID, nil
		}
	}
}

func (s *SimpleTextDocsEnum) advanceTarget(target int) (int, error) {
	if s.seekTo > 0 {
		if _, err := s.in.Seek(s.seekTo, 0); err != nil {
			return 0, err
		}
		s.seekTo = -1
	}

	doc, err := s.readDoc()
	if err != nil {
		return 0, err
	}

	for doc < target {
		doc, err = s.readDoc()
		if err != nil {
			return 0, err
		}
	}

	return doc, nil
}

func (s *SimpleTextDocsEnum) Advance(target int) (int, error) {
	if err := s.AdvanceShallow(target); err != nil {
		return 0, err
	}
	return s.advanceTarget(target)
}

func (s *SimpleTextDocsEnum) SlowAdvance(target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SimpleTextDocsEnum) Cost() int64 {
	return int64(s.cost)
}

func (s *SimpleTextDocsEnum) Freq() (int, error) {
	return s.tf, nil
}

func (s *SimpleTextDocsEnum) NextPosition() (int, error) {
	return -1, nil
}

func (s *SimpleTextDocsEnum) StartOffset() (int, error) {
	return -1, nil
}

func (s *SimpleTextDocsEnum) EndOffset() (int, error) {
	return -1, nil
}

func (s *SimpleTextDocsEnum) GetPayload() ([]byte, error) {
	return []byte{}, nil
}

func (s *SimpleTextDocsEnum) AdvanceShallow(target int) error {
	if target > s.nextSkipDoc {
		if _, err := s.skipReader.SkipTo(target); err != nil {
			return err
		}
		if s.skipReader.getNextSkipDoc() != index.NO_MORE_DOCS {
			s.seekTo = s.skipReader.getNextSkipDocFP()
		}
		s.nextSkipDoc = s.skipReader.getNextSkipDoc()
	}
	return nil
}

func (s *SimpleTextDocsEnum) GetImpacts() (index.Impacts, error) {
	if err := s.AdvanceShallow(s.docID); err != nil {
		return nil, err
	}
	return s.skipReader.getImpacts(), nil
}
