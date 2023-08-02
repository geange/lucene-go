package index

import (
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/version"
)

type SegmentInfo interface {
	GetID() []byte
	Name() string
	Dir() store.Directory
	Files() map[string]struct{}
	FilesNum() int
	MaxDoc() (int, error)
	SetMaxDoc(maxDoc int) error
	SetFiles(files map[string]struct{})
	AddFile(file string) error
	GetVersion() *version.Version
	GetMinVersion() *version.Version
	SetUseCompoundFile(isCompoundFile bool)
	GetUseCompoundFile() bool
	SetDiagnostics(diagnostics map[string]string)
	GetDiagnostics() map[string]string
	PutAttribute(key, value string) string
	GetAttributes() map[string]string
	GetIndexSort() Sort
	NamedForThisSegment(file string) string
	GetCodec() Codec
	SetCodec(codec Codec)
}
