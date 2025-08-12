package lucene50

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/bits-and-blooms/bitset"

	"github.com/geange/lucene-go/core/codecs"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
)

var _ index.LiveDocsFormat = &LiveDocsFormat{}

const (
	LIVE_DOCS_EXTENSION       = "liv"              // extension of live docs
	LIVE_DOCS_CODEC_NAME      = "Lucene50LiveDocs" // codec of live docs
	LIVE_DOCS_VERSION_START   = 0                  // supported version range
	LIVE_DOCS_VERSION_CURRENT = LIVE_DOCS_VERSION_START
)

type LiveDocsFormat struct {
}

func NewLiveDocsFormat() *LiveDocsFormat {
	return &LiveDocsFormat{}
}

func (f *LiveDocsFormat) ReadLiveDocs(ctx context.Context, dir store.Directory, info index.SegmentCommitInfo, context *store.IOContext) (util.Bits, error) {
	gen := info.GetDelGen()
	name := coreIndex.FileNameFromGeneration(info.Info().Name(), LIVE_DOCS_EXTENSION, gen)
	length, err := info.Info().MaxDoc()
	if err != nil {
		return nil, err
	}
	input, err := store.OpenChecksumInput(ctx, dir, name)
	if err != nil {
		return nil, err
	}
	if _, err := codecs.CheckIndexHeader(ctx, input, LIVE_DOCS_CODEC_NAME, LIVE_DOCS_VERSION_START,
		LIVE_DOCS_VERSION_CURRENT, info.Info().GetID(), strconv.FormatInt(gen, 36)); err != nil {
		return nil, err
	}

	data := make([]uint64, bits2words(length))
	for i := range data {
		n, err := input.ReadUint64(ctx)
		if err != nil {
			return nil, err
		}
		data[i] = n
	}
	fbs := bitset.From(data)
	// TODO: if (fbs.length() - fbs.cardinality() != info.getDelCount())
	//fbs.Len()
	return fbs, nil
}

func bits2words(numBits int) int {
	if numBits <= 0 {
		return 0
	}
	// 将位数向上取整到最近的 64 的倍数，然后除以 64
	return int(math.Ceil(float64(numBits) / 64))
}

func (f *LiveDocsFormat) WriteLiveDocs(ctx context.Context, bits util.Bits, dir store.Directory, info index.SegmentCommitInfo, newDelCount int, ioContext *store.IOContext) error {
	gen := info.GetNextDelGen()
	name := coreIndex.FileNameFromGeneration(info.Info().Name(), LIVE_DOCS_EXTENSION, gen)
	delCount := 0
	output, err := dir.CreateOutput(ctx, name)
	if err != nil {
		return err
	}
	if err := codecs.WriteIndexHeader(ctx, output, LIVE_DOCS_CODEC_NAME, LIVE_DOCS_VERSION_CURRENT,
		info.Info().GetID(), strconv.FormatInt(gen, 36)); err != nil {
		return err
	}

	words := bits.Words()
	for _, word := range words {
		if err := output.WriteUint64(ctx, word); err != nil {
			return err
		}
	}
	if err := codecs.WriteFooter(ctx, output); err != nil {
		return err
	}

	//longCount := bits2words(int(bits.Len()))
	//for i := 0; i < longCount; i++ {
	//	currentBits := uint64(0)
	//	j := i << 6
	//	end := min(j+63, int(bits.Len()-1))
	//	for ; j < end; j++ {
	//		if bits.Test(uint(j)) {
	//			currentBits = currentBits | uint64(1<<j)
	//		} else {
	//			currentBits += 1
	//		}
	//	}
	//}

	if delCount != info.GetDelCount()+newDelCount {
		return fmt.Errorf("corrupt index exception, bits.deleted=%d info.delcount=%d newdelcount=%d",
			delCount, info.GetDelCount(), newDelCount)
	}
	return nil
}

func (f *LiveDocsFormat) Files(ctx context.Context, info index.SegmentCommitInfo,
	files map[string]struct{}) (map[string]struct{}, error) {
	if info.HasDeletions() {
		name := coreIndex.FileNameFromGeneration(info.Info().Name(), LIVE_DOCS_EXTENSION, info.GetDelGen())
		files[name] = struct{}{}
	}
	return files, nil
}
