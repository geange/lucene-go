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
