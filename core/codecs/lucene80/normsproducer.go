package lucene80

import (
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
)

var _ index.NormsProducer = &NormsProducer{}

type NormsProducer struct {
}

func (n *NormsProducer) Close() error {
	//TODO implement me
	panic("implement me")
}

func (n *NormsProducer) GetNorms(field *document.FieldInfo) (index.NumericDocValues, error) {
	//TODO implement me
	panic("implement me")
}

func (n *NormsProducer) CheckIntegrity() error {
	//TODO implement me
	panic("implement me")
}

func (n *NormsProducer) GetMergeInstance() index.NormsProducer {
	//TODO implement me
	panic("implement me")
}
