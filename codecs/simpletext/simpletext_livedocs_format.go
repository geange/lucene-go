package simpletext

import (
	"bytes"
	"errors"
	"github.com/bits-and-blooms/bitset"
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

var _ index.LiveDocsFormat = &SimpleTextLiveDocsFormat{}

// SimpleTextLiveDocsFormat reads/writes plaintext live docs
// FOR RECREATIONAL USE ONLY
// lucene.experimental
type SimpleTextLiveDocsFormat struct {
}

func (s *SimpleTextLiveDocsFormat) ReadLiveDocs(dir store.Directory, info index.SegmentCommitInfo, context *store.IOContext) (util.Bits, error) {
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

	value, err := readValue(in, LIVE_DOCS_FORMAT_SIZE, scratch)
	if err != nil {
		return nil, err
	}
	size, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}

	bits := bitset.New(uint(size))

	if err := ReadLine(in, scratch); err != nil {
		return nil, err
	}
	for !bytes.HasPrefix(scratch.Bytes(), LIVE_DOCS_FORMAT_END) {
		scratch.Next(len(LIVE_DOCS_FORMAT_DOC))
		docid, err := strconv.Atoi(scratch.String())
		if err != nil {
			return nil, err
		}
		bits.Set(uint(docid))

		if err := ReadLine(in, scratch); err != nil {
			return nil, err
		}
	}
	if err := CheckFooter(in); err != nil {
		return nil, err
	}
	return NewSimpleTextBits(bits, size), nil
}

func (s *SimpleTextLiveDocsFormat) WriteLiveDocs(bits util.Bits, dir store.Directory, info *index.SegmentCommitInfo, newDelCount int, context *store.IOContext) error {
	size := bits.Length()

	fileName := index.FileNameFromGeneration(info.Info().Name(), LIVEDOCS_EXTENSION, info.GetNextDelGen())

	out, err := dir.CreateOutput(fileName, context)
	if err != nil {
		return err
	}
	if err := writeValue(out, LIVE_DOCS_FORMAT_SIZE, size); err != nil {
		return err
	}

	for i := 0; i < size; i++ {
		if bits.Get(i) {
			if err := writeValue(out, LIVE_DOCS_FORMAT_DOC, size); err != nil {
				return err
			}
		}
	}

	if err := WriteBytes(out, LIVE_DOCS_FORMAT_END); err != nil {
		return err
	}
	if err := WriteNewline(out); err != nil {
		return err
	}
	if err := WriteChecksum(out); err != nil {
		return err
	}
	return nil
}

func (s *SimpleTextLiveDocsFormat) Files(info *index.SegmentCommitInfo, files map[string]struct{}) (map[string]struct{}, error) {
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

func (s *SimpleTextBits) Get(index int) bool {
	return s.bits.Test(uint(index))
}

func (s *SimpleTextBits) Length() int {
	return s.size
}
