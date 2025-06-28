package blocktree

import (
	"context"
	"errors"

	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/fst"
)

var _ index.TermsEnum = &IntersectTermsEnum{}

type IntersectTermsEnum struct {
	*coreIndex.BaseTermsEnum

	in             store.IndexInput
	fstOutputs     *fst.Outputs[[]byte]
	stack          []*IntersectTermsEnumFrame
	arcs           *fst.Arc
	commonSuffix   []byte
	currentFrame   *IntersectTermsEnumFrame
	term           []byte
	fstReader      fst.BytesReader
	fr             *FieldReader
	savedStartTerm []byte
}

func (i *IntersectTermsEnum) Next(ctx context.Context) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (i *IntersectTermsEnum) popPushNext() (bool, error) {
	// Pop finished frames
	panic("")
}

func (i *IntersectTermsEnum) SeekCeil(ctx context.Context, text []byte) (index.SeekStatus, error) {
	return 0, errors.New("unsupported operation exception")
}

func (i *IntersectTermsEnum) SeekExactByOrd(ctx context.Context, ord int64) error {
	return errors.New("unsupported operation exception")
}

func (i *IntersectTermsEnum) Term() ([]byte, error) {
	return i.term, nil
}

func (i *IntersectTermsEnum) Ord() (int64, error) {
	return 0, errors.New("unsupported operation exception")
}

func (i *IntersectTermsEnum) DocFreq() (int, error) {
	if err := i.currentFrame.DecodeMetaData(context.Background()); err != nil {
		return 0, err
	}
	return i.currentFrame.termState.GetDocFreq(), nil
}

func (i *IntersectTermsEnum) TotalTermFreq() (int64, error) {
	if err := i.currentFrame.DecodeMetaData(context.Background()); err != nil {
		return 0, err
	}
	return int64(i.currentFrame.termState.GetTotalTermFreq()), nil
}

func (i *IntersectTermsEnum) Postings(reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {
	if err := i.currentFrame.DecodeMetaData(context.Background()); err != nil {
		return nil, err
	}
	return i.fr.parent.postingsReader.Postings(
		context.Background(), i.fr.fieldInfo, i.currentFrame.termState, reuse, flags)
}

func (i *IntersectTermsEnum) Impacts(flags int) (index.ImpactsEnum, error) {
	if err := i.currentFrame.DecodeMetaData(context.Background()); err != nil {
		return nil, err
	}
	return i.fr.parent.postingsReader.Impacts(context.Background(),
		i.fr.fieldInfo, i.currentFrame.termState, flags)
}
