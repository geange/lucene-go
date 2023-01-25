package index

import (
	"fmt"
	"github.com/geange/lucene-go/core/store"
)

// SortFieldProvider Reads/Writes a named SortField from a segment info file, used to record index sorts
type SortFieldProvider interface {
	NamedSPI

	// ReadSortField Reads a SortField from serialized bytes
	ReadSortField(in store.DataInput) (*SortField, error)

	// WriteSortField Writes a SortField to a DataOutput This is used to record index
	// ort information in segment headers
	WriteSortField(sf *SortField, out store.DataOutput) error
}

var (
	SingleSortFieldProvider = SortFieldProviderInstance{
		values: map[string]SortFieldProvider{},
	}
)

type SortFieldProviderInstance struct {
	values map[string]SortFieldProvider
}

func (s *SortFieldProviderInstance) Register(name string, provider SortFieldProvider) {
	s.values[name] = provider
}

// ForName Looks up a SortFieldProvider by name
func (s *SortFieldProviderInstance) ForName(name string) (SortFieldProvider, bool) {
	provider, ok := s.values[name]
	return provider, ok
}

func (s *SortFieldProviderInstance) MustForName(name string) SortFieldProvider {
	return s.values[name]
}

func (s *SortFieldProviderInstance) Write(sf *SortField, out store.DataOutput) error {
	sorter := sf.GetIndexSorter()
	if sorter != nil {
		return fmt.Errorf("cannot serialize sort field: %s", sf.String())
	}

	provider, ok := s.ForName(sorter.GetProviderName())
	if !ok {
		return fmt.Errorf("provider(%s) not found", sorter.GetProviderName())
	}
	return provider.WriteSortField(sf, out)
}
