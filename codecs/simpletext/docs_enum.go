package simpletext

import (
	"bytes"
	"io"
	"strconv"

	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

var _ index.ImpactsEnum = &DocsEnum{}

type DocsEnum struct {
	inStart      store.IndexInput
	in           store.IndexInput
	omitTF       bool
	docID        int
	tf           int
	scratch      *bytes.Buffer
	scratchUTF16 *bytes.Buffer
	cost         int
	skipReader   *SkipReader
	nextSkipDoc  int
	seekTo       int64
}

func (s *FieldsReader) NewSimpleTextDocsEnum() *DocsEnum {
	return &DocsEnum{
		inStart:      s.in,
		in:           s.in.Clone(),
		omitTF:       false,
		docID:        -1,
		tf:           0,
		scratch:      nil,
		scratchUTF16: nil,
		cost:         0,
		skipReader:   NewSkipReader(s.in.Clone()),
		nextSkipDoc:  0,
		seekTo:       -1,
	}
}

func (s *DocsEnum) CanReuse(in store.IndexInput) bool {
	return in == s.inStart
}

func (s *DocsEnum) DocID() int {
	return s.docID
}

func (s *DocsEnum) NextDoc() (int, error) {
	return s.Advance(s.docID + 1)
}

func (s *DocsEnum) readDoc() (int, error) {
	if s.docID == types.NO_MORE_DOCS {
		return s.docID, nil
	}
	first := true
	termFreq := 0
	var err error
	for {
		lineStart := s.in.GetFilePointer()
		if err := utils.ReadLine(s.in, s.scratch); err != nil {
			return 0, err
		}

		text := s.scratch.Bytes()

		if bytes.HasPrefix(text, FIELDS_DOC) {
			if !first {
				if _, err := s.in.Seek(lineStart, io.SeekStart); err != nil {
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
				if _, err := s.in.Seek(lineStart, io.SeekStart); err != nil {
					return 0, err
				}
				if !s.omitTF {
					s.tf = termFreq
				}
				return s.docID, nil
			}
			s.docID = types.NO_MORE_DOCS
			return s.docID, nil
		}
	}
}

func (s *DocsEnum) advanceTarget(target int) (int, error) {
	if s.seekTo > 0 {
		if _, err := s.in.Seek(s.seekTo, io.SeekStart); err != nil {
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

func (s *DocsEnum) Advance(target int) (int, error) {
	if err := s.AdvanceShallow(target); err != nil {
		return 0, err
	}
	return s.advanceTarget(target)
}

func (s *DocsEnum) SlowAdvance(target int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *DocsEnum) Cost() int64 {
	return int64(s.cost)
}

func (s *DocsEnum) Freq() (int, error) {
	return s.tf, nil
}

func (s *DocsEnum) NextPosition() (int, error) {
	return -1, nil
}

func (s *DocsEnum) StartOffset() (int, error) {
	return -1, nil
}

func (s *DocsEnum) EndOffset() (int, error) {
	return -1, nil
}

func (s *DocsEnum) GetPayload() ([]byte, error) {
	return []byte{}, nil
}

func (s *DocsEnum) AdvanceShallow(target int) error {
	if target > s.nextSkipDoc {
		if _, err := s.skipReader.SkipTo(target); err != nil {
			return err
		}
		if s.skipReader.getNextSkipDoc() != types.NO_MORE_DOCS {
			s.seekTo = s.skipReader.getNextSkipDocFP()
		}
		s.nextSkipDoc = s.skipReader.getNextSkipDoc()
	}
	return nil
}

func (s *DocsEnum) GetImpacts() (index.Impacts, error) {
	if err := s.AdvanceShallow(s.docID); err != nil {
		return nil, err
	}
	return s.skipReader.getImpacts(), nil
}
