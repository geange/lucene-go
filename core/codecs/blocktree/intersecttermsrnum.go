package blocktree

import (
	"context"

	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.TermsEnum = &IntersectTermsEnum{}

type IntersectTermsEnum struct {
	*coreIndex.BaseTermsEnum
}

func (i *IntersectTermsEnum) Next(ctx context.Context) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (i *IntersectTermsEnum) SeekCeil(ctx context.Context, text []byte) (index.SeekStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (i *IntersectTermsEnum) SeekExactByOrd(ctx context.Context, ord int64) error {
	//TODO implement me
	panic("implement me")
}

func (i *IntersectTermsEnum) Term() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (i *IntersectTermsEnum) Ord() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (i *IntersectTermsEnum) DocFreq() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (i *IntersectTermsEnum) TotalTermFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (i *IntersectTermsEnum) Postings(reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (i *IntersectTermsEnum) Impacts(flags int) (index.ImpactsEnum, error) {
	//TODO implement me
	panic("implement me")
}
