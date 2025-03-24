package index

import (
	"iter"

	"github.com/geange/lucene-go/core/interface/index"
)

type FilterLeafReader struct {
}

var _ index.Fields = &FilterFields{}

type FilterFields struct {
	in index.Fields
}

func (f *FilterFields) Iterator() iter.Seq[string] {
	return f.in.Iterator()
}

func (f *FilterFields) Names() []string {
	res := make([]string, 0)
	for v := range f.in.Iterator() {
		res = append(res, v)
	}
	return res
}

func (f *FilterFields) Terms(field string) (index.Terms, error) {
	return f.in.Terms(field)
}

func (f *FilterFields) Size() int {
	return f.in.Size()
}
