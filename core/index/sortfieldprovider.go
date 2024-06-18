package index

import (
	"context"
	"errors"
	"fmt"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

// SortFieldProvider Reads/Writes a named SortField from a segment info file, used to record index sorts
type SortFieldProvider interface {
	Named

	// ReadSortField Reads a SortField from serialized bytes
	ReadSortField(ctx context.Context, in store.DataInput) (index.SortField, error)

	// WriteSortField Writes a SortField to a DataOutput This is used to record index
	// ort information in segment headers
	WriteSortField(ctx context.Context, sf index.SortField, out store.DataOutput) error
}

func RegisterSortFieldProvider(provider SortFieldProvider) {
	sortFieldProviderPool[provider.GetName()] = provider
}

func GetSortFieldProviderByName(name string) SortFieldProvider {
	return sortFieldProviderPool[name]
}

func WriteSortField(sf index.SortField, output store.DataOutput) error {
	sorter := sf.GetIndexSorter()
	if sorter == nil {
		return errors.New("cannot serialize sort field")
	}
	provider := GetSortFieldProviderByName(sorter.GetProviderName())
	if provider != nil {
		return provider.WriteSortField(nil, sf, output)
	}
	return fmt.Errorf("SortFieldProvider: %s not found", sorter.GetProviderName())
}

var (
	sortFieldProviderPool = make(map[string]SortFieldProvider)
)

//type SortFieldProviderInstance struct {
//	values map[string]SortFieldProvider
//}
//
//func (s *SortFieldProviderInstance) Register(name string, provider SortFieldProvider) {
//	s.values[name] = provider
//}
//
//// GetCodecByName Looks up a SortFieldProvider by name
//func (s *SortFieldProviderInstance) GetCodecByName(name string) (SortFieldProvider, bool) {
//	provider, ok := s.values[name]
//	return provider, ok
//}
//
//func (s *SortFieldProviderInstance) MustForName(name string) SortFieldProvider {
//	return s.values[name]
//}
//
//func (s *SortFieldProviderInstance) Write(sf SortField, out store.DataOutput) error {
//	sorter := sf.GetIndexSorter()
//	if sorter != nil {
//		return fmt.Errorf("cannot serialize sort field: %s", sf.String())
//	}
//
//	provider, ok := s.GetCodecByName(sorter.GetProviderName())
//	if !ok {
//		return fmt.Errorf("provider(%s) not found", sorter.GetProviderName())
//	}
//	return provider.WriteSortField(sf, out)
//}
