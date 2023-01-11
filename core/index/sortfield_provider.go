package index

import (
	"github.com/geange/lucene-go/core/store"
)

// SortFieldProvider Reads/Writes a named SortField from a segment info file, used to record index sorts
type SortFieldProvider interface {
	NamedSPI

	// ReadSortField Reads a SortField from serialized bytes
	ReadSortField(in store.DataInput) (*SortField, error)
}

// LooksUpSortFieldProviderByName Looks up a SortFieldProvider by name
func LooksUpSortFieldProviderByName(name string) SortFieldProvider {
	panic("")
}