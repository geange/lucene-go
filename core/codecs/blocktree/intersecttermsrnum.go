package blocktree

import (
	"bytes"
	"context"
	"errors"
	"github.com/geange/lucene-go/core/util/array"

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

// NOTE: specialized to only doing the first-time
// seek, but we could generalize it to allow
// arbitrary seekExact/Ceil.  Note that this is a
// seekFloor!
func (i *IntersectTermsEnum) seekToStartTerm(ctx context.Context, target []byte) error {
	if len(i.term) < len(target) {
		i.term = array.Grow(i.term, len(target))
	}

	for idx := 0; idx < len(target); idx++ {
		for {
			savNextEnt := i.currentFrame.nextEnt
			savePos := i.currentFrame.suffixesReader.GetPosition()
			saveLengthPos := i.currentFrame.suffixLengthsReader.GetPosition()
			saveStartBytePos := i.currentFrame.startBytePos
			saveSuffix := i.currentFrame.suffix
			saveLastSubFP := i.currentFrame.lastSubFP
			saveTermBlockOrd := i.currentFrame.termState.GetTermBlockOrd()

			//isSubBlock := i.currentFrame.Next()

			i.term = array.Grow(i.term, i.currentFrame.prefix+i.currentFrame.suffix)

			size := i.currentFrame.suffix

			copy(i.term[i.currentFrame.prefix:],
				i.currentFrame.suffixBytes[i.currentFrame.startBytePos:i.currentFrame.startBytePos+size])

			cmp := bytes.Compare(i.term, target)
			if cmp < 0 {
				if i.currentFrame.nextEnt == i.currentFrame.entCount {
					if !i.currentFrame.isLastInFloor {
						// Advance to next floor block
						err := i.currentFrame.loadNextFloorBlock(ctx)
						if err != nil {
							return err
						}
						continue
					} else {
						return nil
					}
				}
			} else if cmp == 0 {
				return nil
			} else {
				// Fallback to prior entry: the semantics of
				// this method is that the first call to
				// next() will return the term after the
				// requested term
				i.currentFrame.nextEnt = savNextEnt
				i.currentFrame.lastSubFP = saveLastSubFP
				i.currentFrame.startBytePos = saveStartBytePos
				i.currentFrame.suffix = saveSuffix
				i.currentFrame.suffixesReader.SetPosition(savePos)
				i.currentFrame.suffixLengthsReader.SetPosition(saveLengthPos)
				i.currentFrame.termState.SetTermBlockOrd(saveTermBlockOrd)

				copySize := i.currentFrame.suffix
				srcBytes := i.currentFrame.suffixBytes[i.currentFrame.startBytePos : i.currentFrame.startBytePos+copySize]
				dstBytes := i.term[i.currentFrame.prefix : i.currentFrame.prefix+copySize]
				copy(dstBytes, srcBytes)

				termSize := i.currentFrame.prefix + i.currentFrame.suffix
				i.term = i.term[:termSize]
				// If the last entry was a block we don't
				// need to bother recursing and pushing to
				// the last term under it because the first
				// next() will simply skip the frame anyway
				return nil
			}
		}
	}
	return nil
}
