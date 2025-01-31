package simpletext

import (
	"bytes"
	"context"
	"io"
	"strconv"

	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

var _ index.ImpactsEnum = &PostingsEnum{}

type PostingsEnum struct {
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
	skipReader     *SkipReader
	nextSkipDoc    int
	seekTo         int64
}

func (s *PostingsEnum) DocID() int {
	return s.docID
}

func (s *PostingsEnum) NextDoc(ctx context.Context) (int, error) {
	return s.Advance(ctx, s.docID+1)
}

func (s *PostingsEnum) readDoc() (int, error) {
	first := true
	if _, err := s.in.Seek(s.nextDocStart, io.SeekStart); err != nil {
		return 0, err
	}
	posStart := int64(0)
	var err error
	for {
		lineStart := s.in.GetFilePointer()
		if err := utils.ReadLine(s.in, s.scratch); err != nil {
			return 0, err
		}

		if bytes.HasPrefix(s.scratch.Bytes(), FIELDS_DOC) {
			if !first {
				s.nextDocStart = lineStart
				if _, err := s.in.Seek(posStart, io.SeekStart); err != nil {
					return 0, err
				}
				return s.docID, nil
			}
			s.scratchUTF16.Write(s.scratch.Bytes()[len(FIELDS_DOC):])

			s.docID, err = strconv.Atoi(s.scratchUTF16.String())
			if err != nil {
				return 0, err
			}

			s.tf = 0
			first = false
		} else if bytes.HasPrefix(s.scratch.Bytes(), FIELDS_FREQ) {
			s.scratchUTF16.Write(s.scratch.Bytes()[len(FIELDS_FREQ):])
			s.tf, err = strconv.Atoi(s.scratchUTF16.String())
			posStart = s.in.GetFilePointer()
		} else if bytes.HasPrefix(s.scratch.Bytes(), FIELDS_POS) {
			// skip
		} else if bytes.HasPrefix(s.scratch.Bytes(), FIELDS_START_OFFSET) {
			// skip
		} else if bytes.HasPrefix(s.scratch.Bytes(), FIELDS_END_OFFSET) {
			// skip
		} else if bytes.HasPrefix(s.scratch.Bytes(), FIELDS_PAYLOAD) {
			// skip
		} else {

			if !first {
				s.nextDocStart = lineStart
				if _, err := s.in.Seek(posStart, io.SeekStart); err != nil {
					return 0, err
				}
				return s.docID, nil
			}
			s.docID = types.NO_MORE_DOCS
			return s.docID, nil
		}
	}
}

func (s *PostingsEnum) advanceTarget(target int) (int, error) {
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

func (s *PostingsEnum) Advance(ctx context.Context, target int) (int, error) {
	if err := s.AdvanceShallow(nil, target); err != nil {
		return 0, err
	}
	return s.advanceTarget(target)
}

func (s *PostingsEnum) SlowAdvance(ctx context.Context, target int) (int, error) {
	doc := 0
	var err error
	for doc < target {
		doc, err = s.NextDoc(ctx)
		if err != nil {
			return 0, nil
		}
	}
	return doc, nil
}

func (s *PostingsEnum) Cost() int64 {
	return int64(s.cost)
}

func (s *PostingsEnum) Freq() (int, error) {
	return s.tf, nil
}

func (s *PostingsEnum) NextPosition() (int, error) {
	if s.readPositions {
		if err := utils.ReadLine(s.in, s.scratch); err != nil {
			return 0, err
		}
		s.scratchUTF16_2.Reset()
		s.scratchUTF16_2.Write(s.scratch.Bytes()[len(FIELDS_POS):])
		var err error
		s.pos, err = strconv.Atoi(s.scratchUTF16_2.String())
		if err != nil {
			return 0, err
		}
	} else {
		s.pos = -1
	}

	if s.readOffsets {
		if err := utils.ReadLine(s.in, s.scratch); err != nil {
			return 0, err
		}
		s.scratchUTF16_2.Reset()
		s.scratchUTF16_2.Write(s.scratch.Bytes()[len(FIELDS_START_OFFSET):])
		var err error
		s.startOffset, err = strconv.Atoi(s.scratchUTF16_2.String())
		if err != nil {
			return 0, err
		}

		if err := utils.ReadLine(s.in, s.scratch); err != nil {
			return 0, err
		}
		s.scratchUTF16_2.Reset()
		s.scratchUTF16_2.Write(s.scratch.Bytes()[len(FIELDS_END_OFFSET):])

		s.endOffset, err = strconv.Atoi(s.scratchUTF16_2.String())
		if err != nil {
			return 0, err
		}
	}

	fp := s.in.GetFilePointer()
	if err := utils.ReadLine(s.in, s.scratch); err != nil {
		return 0, err
	}
	if bytes.HasPrefix(s.scratch.Bytes(), FIELDS_PAYLOAD) {
		s.scratch2.Reset()
		s.scratch2.Write(s.scratch.Bytes()[len(FIELDS_PAYLOAD):])
		s.payload = s.scratch2.Bytes()
	} else {
		s.payload = s.payload[:0]
		if _, err := s.in.Seek(fp, io.SeekStart); err != nil {
			return 0, err
		}
	}
	return s.pos, nil
}

func (s *PostingsEnum) StartOffset() (int, error) {
	return s.startOffset, nil
}

func (s *PostingsEnum) EndOffset() (int, error) {
	return s.endOffset, nil
}

func (s *PostingsEnum) GetPayload() ([]byte, error) {
	return s.payload, nil
}

func (s *PostingsEnum) AdvanceShallow(ctx context.Context, target int) error {
	if target > s.nextSkipDoc {
		if _, err := s.skipReader.SkipTo(nil, target); err != nil {
			return err
		}
		if s.skipReader.GetNextSkipDoc() != types.NO_MORE_DOCS {
			s.seekTo = s.skipReader.GetNextSkipDocFP()
		}
	}
	s.nextSkipDoc = s.skipReader.GetNextSkipDoc()
	return nil
}

func (s *PostingsEnum) GetImpacts() (index.Impacts, error) {
	if err := s.AdvanceShallow(nil, s.docID); err != nil {
		return nil, err
	}
	return s.skipReader.GetImpacts(), nil
}
