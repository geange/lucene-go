package simpletext

import (
	"bytes"
	"errors"
	"github.com/bits-and-blooms/bitset"
	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	"strconv"
)

var (
	LIVEDOCS_EXTENSION = "liv"

	LIVE_DOCS_FORMAT_SIZE = []byte("size ")
	LIVE_DOCS_FORMAT_DOC  = []byte("  doc ")
	LIVE_DOCS_FORMAT_END  = []byte("END")
)

var _ index.LiveDocsFormat = &LiveDocsFormat{}

// LiveDocsFormat reads/writes plaintext live docs
// FOR RECREATIONAL USE ONLY
// lucene.experimental
type LiveDocsFormat struct {
}

func NewLiveDocsFormat() *LiveDocsFormat {
	return &LiveDocsFormat{}
}

func (s *LiveDocsFormat) ReadLiveDocs(dir store.Directory, info *index.SegmentCommitInfo, context *store.IOContext) (util.Bits, error) {
	if !info.HasDeletions() {
		return nil, errors.New("hasDeletions")
	}

	fileName := index.FileNameFromGeneration(info.Info().Name(), LIVEDOCS_EXTENSION, info.GetDelGen())

	scratch := new(bytes.Buffer)
	in, err := store.OpenChecksumInput(dir, fileName, context)
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
	return NewSimpleTextBits(bits, size), nil
}

func (s *LiveDocsFormat) WriteLiveDocs(bits util.Bits, dir store.Directory, info *index.SegmentCommitInfo, newDelCount int, context *store.IOContext) error {
	size := int(bits.Len())

	fileName := index.FileNameFromGeneration(info.Info().Name(), LIVEDOCS_EXTENSION, info.GetNextDelGen())

	out, err := dir.CreateOutput(fileName, context)
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
	return nil
}

func (s *LiveDocsFormat) Files(info *index.SegmentCommitInfo, files map[string]struct{}) (map[string]struct{}, error) {
	if info.HasDeletions() {
		fileName := index.FileNameFromGeneration(info.Info().Name(), LIVEDOCS_EXTENSION, info.GetDelGen())
		files[fileName] = struct{}{}
	}
	return files, nil
}

var _ util.Bits = &SimpleTextBits{}

type SimpleTextBits struct {
	bits *bitset.BitSet
	size int
}

func NewSimpleTextBits(bits *bitset.BitSet, size int) *SimpleTextBits {
	return &SimpleTextBits{bits: bits, size: size}
}

func (s *SimpleTextBits) Test(index uint) bool {
	return s.bits.Test(uint(index))
}

func (s *SimpleTextBits) Len() uint {
	return uint(s.size)
}
