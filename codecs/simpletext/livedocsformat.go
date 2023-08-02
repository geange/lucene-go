package simpletext

import (
	"bytes"
	"context"
	"errors"
	"strconv"

	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/codecs/utils"
	coreIndex "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
)

var (
	LIVEDOCS_EXTENSION = "liv"

	LIVE_DOCS_FORMAT_SIZE = []byte("size ")
	LIVE_DOCS_FORMAT_DOC  = []byte("  doc ")
	LIVE_DOCS_FORMAT_END  = []byte("END")
)

var _ index.LiveDocsFormat = &LiveDocsFormat{}

// LiveDocsFormat
// reads/writes plaintext live docs
// FOR RECREATIONAL USE ONLY
// lucene.experimental
type LiveDocsFormat struct {
}

func NewLiveDocsFormat() *LiveDocsFormat {
	return &LiveDocsFormat{}
}

func (s *LiveDocsFormat) ReadLiveDocs(ctx context.Context, dir store.Directory, info *index.SegmentCommitInfo, ioContext *store.IOContext) (util.Bits, error) {
	if !info.HasDeletions() {
		return nil, errors.New("hasDeletions")
	}

	fileName := coreIndex.FileNameFromGeneration(info.Info().Name(), LIVEDOCS_EXTENSION, info.GetDelGen())

	scratch := new(bytes.Buffer)
	in, err := store.OpenChecksumInput(dir, fileName)
	if err != nil {
		return nil, err
	}
	defer in.Close()

	r := utils.NewTextReader(in, scratch)

	value, err := r.ReadLabel(LIVE_DOCS_FORMAT_SIZE)
	if err != nil {
		return nil, err
	}
	size, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}

	bits := bitset.New(uint(size))

	if err := r.ReadLine(); err != nil {
		return nil, err
	}
	for !bytes.HasPrefix(scratch.Bytes(), LIVE_DOCS_FORMAT_END) {
		scratch.Next(len(LIVE_DOCS_FORMAT_DOC))
		docid, err := strconv.Atoi(scratch.String())
		if err != nil {
			return nil, err
		}
		bits.Set(uint(docid))

		if err := r.ReadLine(); err != nil {
			return nil, err
		}
	}
	if err := utils.CheckFooter(in); err != nil {
		return nil, err
	}
	return newSimpleTextBits(bits, size), nil
}

func (s *LiveDocsFormat) WriteLiveDocs(ctx context.Context, bits util.Bits, dir store.Directory, info *index.SegmentCommitInfo, newDelCount int, ioContext *store.IOContext) error {
	size := int(bits.Len())

	fileName := coreIndex.FileNameFromGeneration(info.Info().Name(), LIVEDOCS_EXTENSION, info.GetNextDelGen())

	out, err := dir.CreateOutput(ctx, fileName)
	if err != nil {
		return err
	}
	if err := writeValue(out, LIVE_DOCS_FORMAT_SIZE, size); err != nil {
		return err
	}

	for i := 0; i < size; i++ {
		if bits.Test(uint(i)) {
			if err := writeValue(out, LIVE_DOCS_FORMAT_DOC, size); err != nil {
				return err
			}
		}
	}

	if err := utils.WriteBytes(out, LIVE_DOCS_FORMAT_END); err != nil {
		return err
	}
	if err := utils.NewLine(out); err != nil {
		return err
	}
	if err := utils.WriteChecksum(out); err != nil {
		return err
	}
	return out.Close()
}

func (s *LiveDocsFormat) Files(ctx context.Context, info *index.SegmentCommitInfo, files map[string]struct{}) (map[string]struct{}, error) {
	if info.HasDeletions() {
		fileName := coreIndex.FileNameFromGeneration(info.Info().Name(), LIVEDOCS_EXTENSION, info.GetDelGen())
		files[fileName] = struct{}{}
	}
	return files, nil
}

var _ util.Bits = &simpleTextBits{}

type simpleTextBits struct {
	bits *bitset.BitSet
	size int
}

func newSimpleTextBits(bits *bitset.BitSet, size int) *simpleTextBits {
	return &simpleTextBits{bits: bits, size: size}
}

func (s *simpleTextBits) Test(index uint) bool {
	return s.bits.Test(uint(index))
}

func (s *simpleTextBits) Len() uint {
	return uint(s.size)
}
