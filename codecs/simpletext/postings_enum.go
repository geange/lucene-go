package simpletext

import (
	"bytes"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"strconv"
)

var _ index.ImpactsEnum = &SimpleTextPostingsEnum{}

type SimpleTextPostingsEnum struct {
	inStart        store.IndexInput
	in             store.IndexInput
	docID          int
	tf             int
	scratch        *bytes.Buffer
	scratch2       *bytes.Buffer
	scratchUTF16   *bytes.Buffer
	scratchUTF16_2 *bytes.Buffer
	pos            int
	payload        []byte
	nextDocStart   int64
	readOffsets    bool
	readPositions  bool
	startOffset    int
	endOffset      int
	cost           int
	skipReader     *SimpleTextSkipReader
	nextSkipDoc    int
	seekTo         int64
}

func (s *SimpleTextPostingsEnum) DocID() int {
	return s.docID
}

func (s *SimpleTextPostingsEnum) NextDoc() (int, error) {
	return s.Advance(s.docID + 1)
}

func (s *SimpleTextPostingsEnum) readDoc() (int, error) {
	first := true
	if err := s.in.Seek(s.nextDocStart); err != nil {
		return 0, err
	}
	posStart := int64(0)
	var err error
	for {
		lineStart := s.in.GetFilePointer()
		if err := ReadLine(s.in, s.scratch); err != nil {
			return 0, err
		}
		//System.out.println("NEXT DOC: " + scratch.utf8ToString());
		if bytes.HasPrefix(s.scratch.Bytes(), FieldsToken.DOC) {
			if !first {
				s.nextDocStart = lineStart
				if err := s.in.Seek(posStart); err != nil {
					return 0, err
				}
				return s.docID, nil
			}
			s.scratchUTF16.Write(s.scratch.Bytes()[len(FieldsToken.DOC):])

			s.docID, err = strconv.Atoi(s.scratchUTF16.String())
			if err != nil {
				return 0, err
			}

			s.tf = 0
			first = false
		} else if bytes.HasPrefix(s.scratch.Bytes(), FieldsToken.FREQ) {
			s.scratchUTF16.Write(s.scratch.Bytes()[len(FieldsToken.FREQ):])
			s.tf, err = strconv.Atoi(s.scratchUTF16.String())
			posStart = s.in.GetFilePointer()
		} else if bytes.HasPrefix(s.scratch.Bytes(), FieldsToken.POS) {
			// skip
		} else if bytes.HasPrefix(s.scratch.Bytes(), FieldsToken.START_OFFSET) {
			// skip
		} else if bytes.HasPrefix(s.scratch.Bytes(), FieldsToken.END_OFFSET) {
			// skip
		} else if bytes.HasPrefix(s.scratch.Bytes(), FieldsToken.PAYLOAD) {
			// skip
		} else {

			if !first {
				s.nextDocStart = lineStart
				if err := s.in.Seek(posStart); err != nil {
					return 0, err
				}
				return s.docID, nil
			}
			s.docID = index.NO_MORE_DOCS
			return s.docID, nil
		}
	}
}

func (s *SimpleTextPostingsEnum) advanceTarget(target int) (int, error) {
	if s.seekTo > 0 {
		s.nextDocStart = s.seekTo
		s.seekTo = -1
	}

	var err error
	doc := 0
	doc, err = s.readDoc()
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

func (s *SimpleTextPostingsEnum) Advance(target int) (int, error) {
	if err := s.AdvanceShallow(target); err != nil {
		return 0, err
	}
	return s.advanceTarget(target)
}

func (s *SimpleTextPostingsEnum) SlowAdvance(target int) (int, error) {
	doc := 0
	var err error
	for doc < target {
		doc, err = s.NextDoc()
		if err != nil {
			return 0, nil
		}
	}
	return doc, nil
}

func (s *SimpleTextPostingsEnum) Cost() int64 {
	return int64(s.cost)
}

func (s *SimpleTextPostingsEnum) Freq() (int, error) {
	return s.tf, nil
}

func (s *SimpleTextPostingsEnum) NextPosition() (int, error) {
	if s.readPositions {
		if err := ReadLine(s.in, s.scratch); err != nil {
			return 0, err
		}
		s.scratchUTF16_2.Reset()
		s.scratchUTF16_2.Write(s.scratch.Bytes()[len(FieldsToken.POS):])
		var err error
		s.pos, err = strconv.Atoi(s.scratchUTF16_2.String())
		if err != nil {
			return 0, err
		}
	} else {
		s.pos = -1
	}

	if s.readOffsets {
		if err := ReadLine(s.in, s.scratch); err != nil {
			return 0, err
		}
		s.scratchUTF16_2.Reset()
		s.scratchUTF16_2.Write(s.scratch.Bytes()[len(FieldsToken.START_OFFSET):])
		var err error
		s.startOffset, err = strconv.Atoi(s.scratchUTF16_2.String())
		if err != nil {
			return 0, err
		}

		if err := ReadLine(s.in, s.scratch); err != nil {
			return 0, err
		}
		s.scratchUTF16_2.Reset()
		s.scratchUTF16_2.Write(s.scratch.Bytes()[len(FieldsToken.END_OFFSET):])

		s.endOffset, err = strconv.Atoi(s.scratchUTF16_2.String())
		if err != nil {
			return 0, err
		}
	}

	fp := s.in.GetFilePointer()
	if err := ReadLine(s.in, s.scratch); err != nil {
		return 0, err
	}
	if bytes.HasPrefix(s.scratch.Bytes(), FieldsToken.PAYLOAD) {
		s.scratch2.Reset()
		s.scratch2.Write(s.scratch.Bytes()[len(FieldsToken.PAYLOAD):])
		s.payload = s.scratch2.Bytes()
	} else {
		s.payload = s.payload[:0]
		if err := s.in.Seek(fp); err != nil {
			return 0, err
		}
	}
	return s.pos, nil
}

func (s *SimpleTextPostingsEnum) StartOffset() (int, error) {
	return s.startOffset, nil
}

func (s *SimpleTextPostingsEnum) EndOffset() (int, error) {
	return s.endOffset, nil
}

func (s *SimpleTextPostingsEnum) GetPayload() ([]byte, error) {
	return s.payload, nil
}

func (s *SimpleTextPostingsEnum) AdvanceShallow(target int) error {
	if target > s.nextSkipDoc {
		if _, err := s.skipReader.SkipTo(target); err != nil {
			return err
		}
		if s.skipReader.getNextSkipDoc() != index.NO_MORE_DOCS {
			s.seekTo = s.skipReader.getNextSkipDocFP()
		}
	}
	s.nextSkipDoc = s.skipReader.getNextSkipDoc()
	return nil
}

func (s *SimpleTextPostingsEnum) GetImpacts() (index.Impacts, error) {
	if err := s.AdvanceShallow(s.docID); err != nil {
		return nil, err
	}
	return s.skipReader.getImpacts(), nil
}
