package blocktree

import (
	"bytes"
	"context"

	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/fst"
)

var _ index.TermsEnum = &SegmentTermsEnum{}

func NewSegmentTermsEnum(fr *FieldReader) (*SegmentTermsEnum, error) {
	panic("not implemented")
}

type SegmentTermsEnum struct {
	*coreIndex.BaseTermsEnum

	// Lazy init:
	in store.IndexInput

	stack        []*SegmentTermsEnumFrame
	staticFrame  *SegmentTermsEnumFrame
	currentFrame *SegmentTermsEnumFrame
	termExists   bool
	fr           FieldReader

	targetBeforeCurrentLength int

	//static boolean DEBUG = BlockTreeTermsWriter.DEBUG;

	scratchReader *store.ByteArrayDataInput

	// What prefix of the current term was present in the index; when we only next() through the index, this stays at 0.  It's only set when
	// we seekCeil/Exact:
	validIndexPrefix int

	// assert only:
	eof bool

	term      *bytes.Buffer
	fstReader fst.BytesReader

	arcs []*fst.Arc[[]byte]
}

func (s *SegmentTermsEnum) Next(ctx context.Context) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentTermsEnum) SeekCeil(ctx context.Context, text []byte) (index.SeekStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentTermsEnum) SeekExactByOrd(ctx context.Context, ord int64) error {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentTermsEnum) Term() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentTermsEnum) Ord() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentTermsEnum) DocFreq() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentTermsEnum) TotalTermFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentTermsEnum) Postings(reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentTermsEnum) Impacts(flags int) (index.ImpactsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SegmentTermsEnum) initIndexInput() {
	if s.in == nil {
		s.in = s.fr.parent.termsIn.Clone().(store.IndexInput)
	}
}
