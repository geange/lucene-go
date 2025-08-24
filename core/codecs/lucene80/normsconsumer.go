package lucene80

import (
	"context"
	"errors"
	"io"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.NormsConsumer = &NormsConsumer{}

type NormsConsumer struct {
}

func NewNormsConsumer(ctx context.Context, state *index.SegmentWriteState,
	dataCodec, dataExtension, metaCodec, metaExtension string) (*NormsConsumer, error) {
	panic("")
}

func (n *NormsConsumer) Close() error {
	//TODO implement me
	panic("implement me")
}

func (n *NormsConsumer) AddNormsField(ctx context.Context, field *document.FieldInfo, normsProducer index.NormsProducer) error {
	//TODO implement me
	panic("implement me")
}

func (n *NormsConsumer) Merge(ctx context.Context, mergeState *index.MergeState) error {
	//TODO implement me
	panic("implement me")
}

func (n *NormsConsumer) MergeNormsField(ctx context.Context, mergeFieldInfo *document.FieldInfo, mergeState *index.MergeState) error {
	//TODO implement me
	panic("implement me")
}
func (n *NormsConsumer) writeValues(ctx context.Context, values index.NumericDocValues,
	numBytesPerValue int, out store.IndexOutput) error {

	for {
		if _, err := values.NextDoc(ctx); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		value, err := values.LongValue()
		if err != nil {
			return err
		}
		switch numBytesPerValue {
		case 1:
			if err := out.WriteByte(byte(value)); err != nil {
				return err
			}
		case 2:
			if err := out.WriteUint16(ctx, uint16(value)); err != nil {
				return err
			}
		case 4:
			if err := out.WriteUint32(ctx, uint32(value)); err != nil {
				return err
			}
		case 8:
			if err := out.WriteUint64(ctx, uint64(value)); err != nil {
				return err
			}
		default:
			return errors.New("invalid numBytesPerValue")
		}
	}
}
