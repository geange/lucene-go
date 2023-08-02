package index

import (
	"context"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

type StoredFieldsConsumer struct {
	codec   index.Codec
	dir     store.Directory
	info    *SegmentInfo
	writer  index.StoredFieldsWriter
	lastDoc int
}

func NewStoredFieldsConsumer(codec index.Codec, dir store.Directory, info *SegmentInfo) *StoredFieldsConsumer {
	return &StoredFieldsConsumer{
		codec:   codec,
		dir:     dir,
		info:    info,
		writer:  nil,
		lastDoc: -1,
	}
}

func (s *StoredFieldsConsumer) writeField(ctx context.Context, info *document.FieldInfo, field document.IndexableField) error {
	return s.writer.WriteField(ctx, info, field)
}

func (s *StoredFieldsConsumer) initStoredFieldsWriter(ctx context.Context) error {
	if s.writer == nil {
		writer, err := s.codec.StoredFieldsFormat().FieldsWriter(ctx, s.dir, s.info, nil)
		if err != nil {
			return err
		}
		s.writer = writer
	}
	return nil
}

func (s *StoredFieldsConsumer) StartDocument(ctx context.Context, docID int) error {
	if err := s.initStoredFieldsWriter(ctx); err != nil {
		return err
	}

	s.lastDoc++

	for s.lastDoc < docID {
		if err := s.writer.StartDocument(ctx); err != nil {
			return err
		}
		if err := s.writer.FinishDocument(ctx); err != nil {
			return err
		}
		s.lastDoc++
	}
	return s.writer.StartDocument(ctx)
}

func (s *StoredFieldsConsumer) FinishDocument() error {
	return s.writer.FinishDocument(nil)
}

func (s *StoredFieldsConsumer) Finish(ctx context.Context, maxDoc int) error {
	for s.lastDoc < maxDoc-1 {
		if err := s.StartDocument(ctx, s.lastDoc); err != nil {
			return err
		}
		if err := s.FinishDocument(); err != nil {
			return err
		}
		s.lastDoc++
	}
	return nil
}

func (s *StoredFieldsConsumer) Flush(ctx context.Context, state *index.SegmentWriteState, sortMap *DocMap) error {
	maxDoc, err := state.SegmentInfo.MaxDoc()
	if err != nil {
		return err
	}

	if err := s.writer.Finish(ctx, state.FieldInfos, maxDoc); err != nil {
		return err
	}
	return s.writer.Close()
}
