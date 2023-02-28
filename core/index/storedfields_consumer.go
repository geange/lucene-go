package index

import (
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

type StoredFieldsConsumer struct {
	codec     Codec
	directory store.Directory
	info      *SegmentInfo
	writer    StoredFieldsWriter
	lastDoc   int
}

func NewStoredFieldsConsumer(codec Codec, directory store.Directory, info *SegmentInfo) *StoredFieldsConsumer {
	return &StoredFieldsConsumer{
		codec:     codec,
		directory: directory,
		info:      info,
		writer:    nil,
		lastDoc:   -1,
	}
}

func (s *StoredFieldsConsumer) writeField(info *types.FieldInfo, field types.IndexableField) error {
	return s.writer.WriteField(info, field)
}

func (s *StoredFieldsConsumer) initStoredFieldsWriter() error {
	if s.writer == nil {
		writer, err := s.codec.StoredFieldsFormat().FieldsWriter(s.directory, s.info, nil)
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

	for s.lastDoc+1 < docID {
		if err := s.writer.StartDocument(); err != nil {
			return err
		}
		if err := s.writer.FinishDocument(); err != nil {
			return err
		}
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
