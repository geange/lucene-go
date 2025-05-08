package lucene84

import (
	"context"

	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

type SkipReader struct {
	sr  *skipReader
	mrx *index.MultiLevelSkipListReaderContext
}

func NewSkipReader(skipStream store.IndexInput, maxSkipLevels int,
	hasPos, hasOffsets, hasPayloads bool) (*SkipReader, error) {
	mrx := index.NewMultiLevelSkipListReaderContext(skipStream, maxSkipLevels, BLOCK_SIZE, 8)
	sr := newSkipReader(skipStream, maxSkipLevels, hasPos, hasOffsets, hasPayloads)
	return &SkipReader{sr: sr, mrx: mrx}, nil
}

var _ index.MultiLevelSkipListReaderSPI = &skipReader{}

type skipReader struct {
	docPointer          []uint64
	posPointer          []uint64
	payPointer          []uint64
	posBufferUpto       []uint64
	payloadByteUpto     []uint64
	lastPosPointer      int64
	lastPayPointer      int64
	lastPayloadByteUpto int
	lastDocPointer      int64
	lastPosBufferUpto   int
}

func newSkipReader(skipStream store.IndexInput, maxSkipLevels int,
	hasPos, hasOffsets, hasPayloads bool) *skipReader {
	sr := &skipReader{}
	sr.docPointer = make([]uint64, maxSkipLevels)
	if hasPos {
		sr.posPointer = make([]uint64, maxSkipLevels)
		sr.posBufferUpto = make([]uint64, maxSkipLevels)
		if hasPayloads {
			sr.payloadByteUpto = make([]uint64, maxSkipLevels)
		} else {
			sr.payloadByteUpto = nil
		}
		if hasOffsets || hasPayloads {
			sr.payPointer = make([]uint64, maxSkipLevels)
		} else {
			sr.payPointer = nil
		}
	} else {
		sr.posPointer = nil
	}
	return sr
}

func (s *SkipReader) GetDocPointer() int64 {
	return s.sr.lastDocPointer
}

func (s *SkipReader) GetPosPointer() int64 {
	return s.sr.lastPosPointer
}

func (s *SkipReader) GetPosBufferUpto() int {
	return s.sr.lastPosBufferUpto
}

func (s *SkipReader) GetPayPointer() int64 {
	return s.sr.lastPayPointer
}

func (s *SkipReader) GetPayloadByteUpto() int {
	return s.sr.lastPayloadByteUpto
}

func (s *SkipReader) GetNextSkipDoc() int {
	return s.mrx.GetSkipDoc(0)
}

func trim(df int) int {
	if df%BLOCK_SIZE == 0 {
		return df - 1
	}
	return df
}

func (s *SkipReader) Init(ctx context.Context, skipPointer, docBasePointer, posBasePointer, payBasePointer int, df int) error {
	if err := s.mrx.Init(ctx, int64(skipPointer), trim(df), s.sr); err != nil {
		return err
	}

	s.sr.lastDocPointer = int64(docBasePointer)
	s.sr.lastPosPointer = int64(posBasePointer)
	s.sr.lastPayPointer = int64(payBasePointer)

	arrayFill(s.sr.docPointer, uint64(docBasePointer))

	if len(s.sr.posPointer) > 0 {
		arrayFill(s.sr.payPointer, uint64(posBasePointer))

		if len(s.sr.payPointer) > 0 {
			arrayFill(s.sr.payPointer, uint64(payBasePointer))
		}
	}
	return nil
}

func (s *SkipReader) SkipTo(ctx context.Context, target int) (int, error) {
	return s.sr.SkipTo(ctx, target, s.mrx)
}

func (s *skipReader) SkipTo(ctx context.Context, target int, mtx *index.MultiLevelSkipListReaderContext) (int, error) {
	return mtx.SkipTo(ctx, target, s, index.WithSeekChild(s.SeekChild))
}

func (s *skipReader) SeekChild(ctx context.Context,
	mrx *index.MultiLevelSkipListReaderContext, level int, spi index.MultiLevelSkipListReaderSPI) error {

	if err := index.DefaultSeekChild(ctx, mrx, level, spi); err != nil {
		return err
	}

	s.docPointer[level] = uint64(s.lastDocPointer)
	if len(s.posPointer) > 0 {
		s.posPointer[level] = uint64(s.lastPosPointer)
		s.posBufferUpto[level] = uint64(s.lastPosBufferUpto)
		if len(s.payloadByteUpto) != 0 {
			s.payloadByteUpto[level] = uint64(s.lastPayloadByteUpto)
		}
		if len(s.payPointer) != 0 {
			s.payPointer[level] = uint64(s.lastPayPointer)
		}
	}
	return nil
}

func (s *SkipReader) GetDoc() int {
	return s.mrx.GetDoc()
}

func (s *skipReader) ReadSkipData(ctx context.Context, level int, skipStream store.IndexInput, mrx *index.MultiLevelSkipListReaderContext) (int64, error) {
	delta, err := skipStream.ReadUvarint(ctx)
	if err != nil {
		return 0, err
	}
	pointer, err := skipStream.ReadUvarint(ctx)
	if err != nil {
		return 0, err
	}
	s.docPointer[level] += pointer

	if s.posPointer != nil {
		p, err := skipStream.ReadUvarint(ctx)
		if err != nil {
			return 0, err
		}
		s.posPointer[level] += p

		upto, err := skipStream.ReadUvarint(ctx)
		if err != nil {
			return 0, err
		}
		s.posBufferUpto[level] = upto

		if s.payloadByteUpto != nil {
			upto, err := skipStream.ReadUvarint(ctx)
			if err != nil {
				return 0, err
			}

			s.payloadByteUpto[level] = upto
		}

		if s.payPointer != nil {
			p, err := skipStream.ReadUvarint(ctx)
			if err != nil {
				return 0, err
			}
			s.payPointer[level] += p
		}
	}
	if err := readImpacts(ctx, level, skipStream); err != nil {
		return 0, err
	}
	return int64(delta), nil
}

// The default impl skips impacts
func readImpacts(ctx context.Context, level int, skipStream store.IndexInput) error {
	n, err := skipStream.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	return skipStream.SkipBytes(ctx, int(n))
}

func (s *skipReader) ReadLevelLength(ctx context.Context, skipStream store.IndexInput, mrx *index.MultiLevelSkipListReaderContext) (int64, error) {
	n, err := skipStream.ReadUvarint(ctx)
	return int64(n), err
}

func (s *skipReader) ReadChildPointer(ctx context.Context, skipStream store.IndexInput, mrx *index.MultiLevelSkipListReaderContext) (int64, error) {
	n, err := skipStream.ReadUvarint(ctx)
	return int64(n), err
}
