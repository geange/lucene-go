package index

import (
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/store"
)

type StoredFieldsConsumer struct {
	codec   Codec
	dir     store.Directory
	info    *SegmentInfo
	writer  StoredFieldsWriter
	lastDoc int
}

func NewStoredFieldsConsumer(codec Codec, dir store.Directory, info *SegmentInfo) *StoredFieldsConsumer {
	return &StoredFieldsConsumer{
		codec:   codec,
		dir:     dir,
		info:    info,
		writer:  nil,
		lastDoc: -1,
	}
}

func (s *StoredFieldsConsumer) writeField(info *document.FieldInfo, field document.IndexableField) error {
	return s.writer.WriteField(info, field)
}

func (s *StoredFieldsConsumer) initStoredFieldsWriter() error {
	if s.writer == nil {
		writer, err := s.codec.StoredFieldsFormat().FieldsWriter(s.dir, s.info, nil)
		if err != nil {
			return err
		}
		s.writer = writer
	}
	return nil
}

func (s *StoredFieldsConsumer) StartDocument(docID int) error {
	if err := s.initStoredFieldsWriter(); err != nil {
		return err
	}

	s.lastDoc++

	for s.lastDoc < docID {
		if err := s.writer.StartDocument(); err != nil {
			return err
		}
		if err := s.writer.FinishDocument(); err != nil {
			return err
		}
		s.lastDoc++
	}
	return s.writer.StartDocument()
}

func (s *StoredFieldsConsumer) FinishDocument() error {
	return s.writer.FinishDocument()
}

func (s *StoredFieldsConsumer) Finish(maxDoc int) error {
	for s.lastDoc < maxDoc-1 {
		if err := s.StartDocument(s.lastDoc); err != nil {
			return err
		}
		if err := s.FinishDocument(); err != nil {
			return err
		}
		s.lastDoc++
	}
	return nil
}

func (s *StoredFieldsConsumer) Flush(state *SegmentWriteState, sortMap *DocMap) error {
	maxDoc, err := state.SegmentInfo.MaxDoc()
	if err != nil {
		return err
	}
	err = s.writer.Finish(state.FieldInfos, maxDoc)
	if err != nil {
		return err
	}
	return s.writer.Close()
}
