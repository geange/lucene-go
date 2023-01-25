package search

import (
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
)

// SortField Stores information about how to sort documents by terms in an individual field.
// Fields must be indexed in order to sort by them.
// Created: Feb 11, 2004 1:25:29 PM
// Since: lucene 1.4
// See Also: Sort
type SortField struct {
}

type SortFieldType int

var _ index.SortFieldProvider = &Provider{}

// Provider A SortFieldProvider for field sorts
type Provider struct {
	name string
}

func (p *Provider) GetName() string {
	return p.name
}

func (p *Provider) ReadSortField(in store.DataInput) (*index.SortField, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Provider) WriteSortField(sf *index.SortField, out store.DataOutput) error {
	//TODO implement me
	panic("implement me")
}
