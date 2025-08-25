package lucene80

import (
	"context"
	"errors"
	"io"
	"math"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.NormsConsumer = &NormsConsumer{}

type NormsConsumer struct {
	data   store.IndexOutput
	meta   store.IndexOutput
	maxDoc int
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
	values, err := normsProducer.GetNorms(field)
	if err != nil {
		return err
	}
	numDocsWithValue := 0
	minV := int64(math.MinInt32)
	maxV := int64(math.MaxInt32)
	for {
		_, err := values.NextDoc(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		numDocsWithValue++
		v, err := values.LongValue()
		if err != nil {
			return err
		}
		minV = min(min, v)
		maxV = max(max, v)
	}

	if err := n.meta.WriteUint32(ctx, uint32(field.Number())); err != nil {
		return err
	}

	if numDocsWithValue == 0 {
		// docsWithFieldOffset
		if err := n.meta.WriteUint64(ctx, uint64(int64(-2))); err != nil {
			return err
		}

		// docsWithFieldLength
		if err := n.meta.WriteUint64(ctx, 0); err != nil {
			return err
		}

		// jumpTableEntryCount
		if err := n.meta.WriteUint16(ctx, uint16(int16(-1))); err != nil {
			return err
		}

		// denseRankPower
		if err := n.meta.WriteByte(uint8(int8(-1))); err != nil {
			return err
		}
	} else if numDocsWithValue == n.maxDoc {
		// docsWithFieldOffset
		if err := n.meta.WriteUint64(ctx, uint64(int64(-1))); err != nil {
			return err
		}

		// docsWithFieldLength
		if err := n.meta.WriteUint64(ctx, 0); err != nil {
			return err
		}

		// jumpTableEntryCount
		if err := n.meta.WriteUint16(ctx, uint16(int16(-1))); err != nil {
			return err
		}

		// denseRankPower
		if err := n.meta.WriteByte(uint8(int8(-1))); err != nil {
			return err
		}
	} else {
		offset := n.data.GetFilePointer()

		// docsWithFieldOffset
		if err := n.meta.WriteUint64(ctx, uint64(offset)); err != nil {
			return err
		}
		values, err = normsProducer.GetNorms(field)

		if err != nil {
			return err
		}
		jumpTableEntryCount, err := WriteBitSet(ctx, values, n.data, DEFAULT_DENSE_RANK_POWER)
		if err != nil {
			return err
		}

		// docsWithFieldLength
		if err := n.meta.WriteUint64(ctx, uint64(n.data.GetFilePointer()-offset)); err != nil {
			return err
		}
		if err := n.meta.WriteUint16(ctx, jumpTableEntryCount); err != nil {
			return err

		}
		if err := n.meta.WriteByte(DEFAULT_DENSE_RANK_POWER); err != nil {
			return err
		}
	}

	if err := n.meta.WriteUint32(ctx, uint32(numDocsWithValue)); err != nil {
		return err
	}
	numBytesPerValue := n.numBytesPerValue(minV, maxV)

	if err := n.meta.WriteByte(byte(numBytesPerValue)); err != nil {
		return err
	}
	if numBytesPerValue == 0 {
		if err := n.meta.WriteUint64(ctx, uint64(minV)); err != nil {
			return err
		}
	} else {
		if err := n.meta.WriteUint64(ctx, uint64(n.data.GetFilePointer())); err != nil { // normsOffset
			return err
		}
		values, err = normsProducer.GetNorms(field)
		if err != nil {
			return err
		}
		if err := n.writeValues(ctx, values, int(numBytesPerValue), n.data); err != nil {
			return err
		}
	}

	return nil
}

func (n *NormsConsumer) Merge(ctx context.Context, mergeState *index.MergeState) error {
	//TODO implement me
	panic("implement me")
}

func (n *NormsConsumer) MergeNormsField(ctx context.Context, mergeFieldInfo *document.FieldInfo, mergeState *index.MergeState) error {
	//TODO implement me
	panic("implement me")
}

func (n *NormsConsumer) numBytesPerValue(minValue, maxValue int64) int64 {
	if minValue >= maxValue {
		return 0
	}

	if minValue >= math.MinInt8 && maxValue <= math.MaxInt8 {
		return 1
	}

	if minValue >= math.MinInt16 && maxValue <= math.MaxInt16 {
		return 2
	}

	if minValue >= math.MinInt32 && maxValue <= math.MaxInt32 {
		return 4
	}

	return 8
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
