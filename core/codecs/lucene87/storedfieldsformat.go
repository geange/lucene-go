package lucene87

import (
	"context"
	"errors"
	"fmt"

	"github.com/geange/lucene-go/core/codecs/compressing"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.StoredFieldsFormat = &StoredFieldsFormat{}

type Mode int // Configuration option for stored fields.

func (m Mode) String() string {
	switch m {
	case BEST_SPEED:
		return "BEST_SPEED"
	case BEST_COMPRESSION:
		return "BEST_COMPRESSION"
	default:
		return ""
	}
}

const (
	BEST_SPEED       = Mode(iota) // Trade compression ratio for retrieval speed.
	BEST_COMPRESSION              // Trade retrieval speed for compression ratio.
	UNKNOWN_MODE
)

func stringToMode(name string) Mode {
	switch name {
	case "BEST_SPEED":
		return BEST_SPEED
	case "BEST_COMPRESSION":
		return BEST_COMPRESSION
	default:
		return UNKNOWN_MODE
	}
}

const (
	MODE_KEY = "Lucene87StoredFieldsFormat.mode"
)

type StoredFieldsFormat struct {
	mode Mode
}

func NewStoredFieldsFormat(mode Mode) *StoredFieldsFormat {
	return &StoredFieldsFormat{mode: mode}
}

func (s *StoredFieldsFormat) FieldsReader(ctx context.Context, directory store.Directory, si index.SegmentInfo, fn index.FieldInfos, ioContext *store.IOContext) (index.StoredFieldsReader, error) {
	value, ok := si.GetAttributes()[MODE_KEY]
	if !ok {
		return nil, fmt.Errorf("missing value for %s for segment: %s", MODE_KEY, si.Name())
	}
	mode := stringToMode(value)
	fieldsFormat, err := s.impl(mode)
	if err != nil {
		return nil, err
	}
	return fieldsFormat.FieldsReader(ctx, directory, si, fn, ioContext)
}

func (s *StoredFieldsFormat) FieldsWriter(ctx context.Context, directory store.Directory, si index.SegmentInfo, ioContext *store.IOContext) (index.StoredFieldsWriter, error) {
	previous := si.PutAttribute(MODE_KEY, s.mode.String())
	if previous != "" && previous != s.mode.String() {
		return nil, fmt.Errorf("found existing value for %s for segment: %s old=%s, new=%s",
			MODE_KEY, si.Name(), previous, s.mode.String())
	}
	format, err := s.impl(s.mode)
	if err != nil {
		return nil, err
	}
	return format.FieldsWriter(ctx, directory, si, ioContext)
}

func (s *StoredFieldsFormat) impl(mode Mode) (index.StoredFieldsFormat, error) {
	switch mode {
	case BEST_SPEED:
		return compressing.NewStoredFieldsFormat("Lucene87StoredFieldsFastData", "", BEST_SPEED_MODE, BEST_SPEED_BLOCK_LENGTH, 1024, 10)
	case BEST_COMPRESSION:
		return compressing.NewStoredFieldsFormat("Lucene87StoredFieldsHighData", "", BEST_COMPRESSION_MODE, BEST_COMPRESSION_BLOCK_LENGTH, 4096, 10)
	default:
		return nil, errors.New("")
	}
}

var (
	BEST_SPEED_MODE       compressing.CompressionMode = NewDeflateWithPresetDictCompressionMode()
	BEST_COMPRESSION_MODE compressing.CompressionMode = NewLZ4WithPresetDictCompressionMode()

	BEST_COMPRESSION_BLOCK_LENGTH = 10 * 48 * 1024
	BEST_SPEED_BLOCK_LENGTH       = 10 * 8 * 1024
)
