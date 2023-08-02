package index

import (
	"context"
	"errors"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

type Named interface {
	GetName() string
}

var codesPool = make(map[string]index.Codec)

func RegisterCodec(codec index.Codec) {
	codesPool[codec.GetName()] = codec
}

func GetCodecByName(name string) (index.Codec, bool) {
	codec, exist := codesPool[name]
	return codec, exist
}

type BaseCompoundDirectory struct {
}

var (
	ErrUnsupportedOperation = errors.New("unsupported operation exception")
)

func (*BaseCompoundDirectory) DeleteFile(ctx context.Context, name string) error {
	return ErrUnsupportedOperation
}

func (*BaseCompoundDirectory) Rename(ctx context.Context, source, dest string) error {
	return ErrUnsupportedOperation
}

func (*BaseCompoundDirectory) SyncMetaData(ctx context.Context) error {
	return nil
}

func (*BaseCompoundDirectory) CreateOutput(ctx context.Context, name string) (store.IndexOutput, error) {
	return nil, ErrUnsupportedOperation
}

func (*BaseCompoundDirectory) CreateTempOutput(ctx context.Context, prefix, suffix string) (store.IndexOutput, error) {
	return nil, ErrUnsupportedOperation
}

func (*BaseCompoundDirectory) Sync(names map[string]struct{}) error {
	return ErrUnsupportedOperation
}

func (*BaseCompoundDirectory) ObtainLock(name string) (store.Lock, error) {
	return nil, ErrUnsupportedOperation
}
