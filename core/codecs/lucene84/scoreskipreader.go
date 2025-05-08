package lucene84

import (
	"context"
	"math"
	"slices"

	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

type ScoreSkipReader struct {
	sr  *scoreSkipReader
	mrx *coreIndex.MultiLevelSkipListReaderContext
}

func NewScoreSkipReader(skipStream store.IndexInput,
	maxSkipLevels int, hasPos, hasOffsets, hasPayloads bool) *ScoreSkipReader {
	mrx := coreIndex.NewMultiLevelSkipListReaderContext(skipStream, maxSkipLevels, BLOCK_SIZE, 8)
	sr := newScoreSkipReader(skipStream, maxSkipLevels, hasPos, hasOffsets, hasPayloads)
	return &ScoreSkipReader{sr: sr, mrx: mrx}
}

func (r *ScoreSkipReader) GetImpacts() (index.Impacts, error) {
	return r.sr.impacts, nil
}

func (r *ScoreSkipReader) SkipTo(ctx context.Context, target int) (int, error) {
	result, err := r.sr.SkipTo(ctx, target, r.mrx)
	if err != nil {
		return 0, err
	}
	if r.mrx.NumberOfSkipLevels() > 0 {
		r.sr.numLevels = r.mrx.NumberOfSkipLevels()
	} else {
		// End of postings don't have skip data anymore, so we fill with dummy data
		// like SlowImpactsEnum.
		r.sr.numLevels = 1
		r.sr.perLevelImpacts[0].SetLength(1)
		r.sr.perLevelImpacts[0].Get(0).SetFreq(math.MaxInt32)
		r.sr.perLevelImpacts[0].impacts[0].SetNorm(1)
		r.sr.impactDataLength[0] = 0
	}
	return result, nil
}

func (r *ScoreSkipReader) GetPosPointer() int64 {
	return r.sr.lastPosPointer
}

func (r *ScoreSkipReader) GetPosBufferUpto() int {
	return r.sr.lastPosBufferUpto
}

func (r *ScoreSkipReader) GetDoc() uint64 {
	return uint64(r.mrx.GetDoc())
}

func (r *ScoreSkipReader) GetDocPointer() int64 {
	return r.sr.lastDocPointer
}

func (r *ScoreSkipReader) GetNextSkipDoc() int {
	return r.mrx.GetSkipDoc(0)
}

func (r *ScoreSkipReader) Init(ctx context.Context, skipPointer, docBasePointer, posBasePointer, payBasePointer int, df int) error {
	if err := r.mrx.Init(ctx, int64(skipPointer), trim(df), r.sr); err != nil {
		return err
	}

	r.sr.lastDocPointer = int64(docBasePointer)
	r.sr.lastPosPointer = int64(posBasePointer)
	r.sr.lastPayPointer = int64(payBasePointer)

	arrayFill(r.sr.docPointer, uint64(docBasePointer))

	if len(r.sr.posPointer) > 0 {
		arrayFill(r.sr.payPointer, uint64(posBasePointer))

		if len(r.sr.payPointer) > 0 {
			arrayFill(r.sr.payPointer, uint64(payBasePointer))
		}
	}
	return nil
}

var _ index.Impacts = &impact{}

type impact struct {
	sr  *scoreSkipReader
	mrx *coreIndex.MultiLevelSkipListReaderContext
}

func newImpact(sr *scoreSkipReader, mrx *coreIndex.MultiLevelSkipListReaderContext) *impact {
	return &impact{
		sr:  sr,
		mrx: mrx,
	}
}

func (i *impact) NumLevels() int {
	return i.sr.numLevels
}

func (i *impact) GetDocIdUpTo(level int) int {
	return i.mrx.GetSkipDoc(level)
}

func (i *impact) GetImpacts(level int) []index.Impact {
	if i.sr.impactDataLength[level] > 0 {
		size := i.sr.impactDataLength[level]
		i.sr.badi.Reset(i.sr.impactData[level][0:size])
		impacts, err := readImpactList(context.Background(), i.sr.badi, i.sr.perLevelImpacts[level])
		if err != nil {
			panic(err)
		}
		i.sr.perLevelImpacts[level] = impacts
		i.sr.impactDataLength[level] = 0
	}
	return i.sr.perLevelImpacts[level].impacts
}

func readImpactList(ctx context.Context, in *store.ByteArrayDataInput, reuse *MutableImpactList) (*MutableImpactList, error) {
	maxNumImpacts := in.Length() // at most one impact per byte
	if reuse.Size() < int(maxNumImpacts) {
		oldLength := reuse.Size()
		reuse.impacts = slices.Grow(reuse.impacts, int(maxNumImpacts))
		size := reuse.Size()
		for i := oldLength; i < size; i++ {
			reuse.impacts[i] = coreIndex.NewImpact(math.MaxInt32, 1)
		}
	}

	freq := 0
	norm := 0
	length := 0
	for in.GetPosition() < int(in.Length()) {
		freqDelta, err := in.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		if (freqDelta & 0x01) != 0 {
			freq += 1 + int(freqDelta>>1)
			zint, err := in.ReadZInt64(ctx)
			if err != nil {
				return nil, err
			}
			norm += 1 + int(zint)
		} else {
			freq += 1 + int(freqDelta>>1)
			norm++
		}

		impact := reuse.impacts[length]
		impact.SetFreq(freq)
		impact.SetNorm(int64(norm))
		length++
	}
	reuse.SetLength(length)
	return reuse, nil
}

var _ coreIndex.MultiLevelSkipListReaderSPI = &scoreSkipReader{}

type scoreSkipReader struct {
	*skipReader

	impactData       [][]byte
	impactDataLength []int
	badi             *store.ByteArrayDataInput
	impacts          index.Impacts
	numLevels        int
	perLevelImpacts  []*MutableImpactList
}

func newScoreSkipReader(skipStream store.IndexInput,
	maxSkipLevels int, hasPos, hasOffsets, hasPayloads bool) *scoreSkipReader {
	this := &scoreSkipReader{
		skipReader:       newSkipReader(skipStream, maxSkipLevels, hasPos, hasOffsets, hasPayloads),
		impactData:       make([][]byte, maxSkipLevels),
		impactDataLength: make([]int, maxSkipLevels),
		perLevelImpacts:  make([]*MutableImpactList, maxSkipLevels),
	}

	for i := range this.impactData {
		this.impactData[i] = []byte{0}
	}

	for i := range this.perLevelImpacts {
		this.perLevelImpacts[i] = newMutableImpactList()
	}
	return this
}

func newMutableImpactList() *MutableImpactList {
	return &MutableImpactList{
		impacts: make([]index.Impact, 0),
	}
}

type MutableImpactList struct {
	impacts []index.Impact
}

func (m *MutableImpactList) SetLength(size int) {
	m.impacts = m.impacts[:size]
}

func (m *MutableImpactList) Get(idx int) index.Impact {
	return m.impacts[idx]
}

func (m *MutableImpactList) Size() int {
	return len(m.impacts)
}
