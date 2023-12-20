package bkd

import (
	"context"
	"fmt"
	"slices"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

func WriteDocIds(ctx context.Context, docIds []int, out store.DataOutput) error {
	// docs can be sorted either when all docs in a block have the same value
	// or when a segment is sorted
	// 已经排序的使用varint写入
	if slices.IsSorted(docIds) {
		if err := out.WriteByte(0); err != nil {
			return err
		}
		previous := 0
		for _, doc := range docIds {
			if err := out.WriteUvarint(ctx, uint64(doc-previous)); err != nil {
				return err
			}
			previous = doc
		}
		return nil
	}

	maxID := 0
	for _, docId := range docIds {
		maxID |= docId
	}

	if maxID <= 0xFFFFFF {
		//
		if err := out.WriteByte(24); err != nil {
			return err
		}

		for _, docId := range docIds {
			if err := out.WriteUint16(ctx, uint16(docId>>8)); err != nil {
				return err
			}
			if err := out.WriteByte(byte(docId)); err != nil {
				return err
			}
		}
		return nil
	}

	// 32位int写入
	if err := out.WriteByte(32); err != nil {
		return err
	}
	for _, docId := range docIds {
		if err := out.WriteUint32(ctx, uint32(docId)); err != nil {
			return err
		}
	}
	return nil
}

// ReadInts Read count integers into docIDs.
func ReadInts(in store.DataInput, count int, docIDs []int) error {
	bpv, err := in.ReadByte()
	if err != nil {
		return err
	}
	switch bpv {
	case 0:
		return readDeltaVInts(in, count, docIDs)
	case 24:
		return readInts24(in, count, docIDs)
	case 32:
		return readInts32(in, count, docIDs)
	default:
		return fmt.Errorf("unsupported number of bits per value: %d", bpv)
	}
}

func readDeltaVInts(in store.DataInput, count int, docIDs []int) error {
	doc := 0
	for i := 0; i < count; i++ {
		value, err := in.ReadUvarint(nil)
		if err != nil {
			return err
		}
		doc += int(value)
		docIDs[i] = doc
	}
	return nil
}

func readInts32(in store.DataInput, count int, docIDs []int) error {
	for i := 0; i < count; i++ {
		v, err := in.ReadUint32(nil)
		if err != nil {
			return err
		}
		docIDs[i] = int(v)
	}
	return nil
}

func readInts24(in store.DataInput, count int, docIDs []int) error {
	i := 0

	bs := make([]byte, 24)

	for ; i < count-7; i += 8 {
		_, err := in.Read(bs)
		if err != nil {
			return err
		}

		docIDs[i] = uint24(bs[:3])
		docIDs[i+1] = uint24(bs[3:6])
		docIDs[i+2] = uint24(bs[6:9])
		docIDs[i+3] = uint24(bs[9:12])
		docIDs[i+4] = uint24(bs[12:15])
		docIDs[i+5] = uint24(bs[15:18])
		docIDs[i+6] = uint24(bs[18:21])
		docIDs[i+7] = uint24(bs[21:24])
	}

	for ; i < count; i++ {
		_, err := in.Read(bs[:3])
		if err != nil {
			return err
		}
		docIDs[i] = uint24(bs[:3])
	}

	return nil
}

func uint24(bs []byte) int {
	return int(bs[0])<<16 | int(bs[1])<<8 | int(bs[2])
}

func putUint24(num int) []byte {
	bs := make([]byte, 3)
	bs[0] = byte(num >> 16)
	bs[1] = byte(num >> 8)
	bs[2] = byte(num)
	return bs
}

// ReadIntsVisitor
// Read count integers and feed the result directly to org.apache.lucene.index.PointValues.IntersectVisitor.visit(int).
func ReadIntsVisitor(ctx context.Context, in store.IndexInput, count int, visitor types.IntersectVisitor) error {
	bpv, err := in.ReadByte()
	if err != nil {
		return err
	}
	switch bpv {
	case 0:
		return readDeltaVIntsVisitor(ctx, in, count, visitor)
	case 32:
		return readInts32Visitor(in, count, visitor)
	case 24:
		return readInts24Visitor(in, count, visitor)
	default:
		return fmt.Errorf("unsupported number of bits per value: %d", bpv)
	}
}

func readDeltaVIntsVisitor(ctx context.Context, in store.IndexInput, count int, visitor types.IntersectVisitor) error {
	doc := 0
	for i := 0; i < count; i++ {
		v, err := in.ReadUvarint(ctx)
		if err != nil {
			return err
		}

		doc += int(v)

		err = visitor.Visit(doc)
		if err != nil {
			return err
		}
	}
	return nil
}

func readInts32Visitor(in store.IndexInput, count int, visitor types.IntersectVisitor) error {
	for i := 0; i < count; i++ {
		v, err := in.ReadUint32(nil)
		if err != nil {
			return err
		}

		err = visitor.Visit(int(v))
		if err != nil {
			return err
		}
	}
	return nil
}

func readInts24Visitor(in store.IndexInput, count int, visitor types.IntersectVisitor) error {
	i := 0

	bs := make([]byte, 24)

	for ; i < count-7; i += 8 {
		_, err := in.Read(bs)
		if err != nil {
			return err
		}

		if err := visitor.Visit(uint24(bs[:3])); err != nil {
			return err
		}
		if err := visitor.Visit(uint24(bs[3:6])); err != nil {
			return err
		}
		if err := visitor.Visit(uint24(bs[6:6])); err != nil {
			return err
		}
		if err := visitor.Visit(uint24(bs[9:12])); err != nil {
			return err
		}
		if err := visitor.Visit(uint24(bs[12:15])); err != nil {
			return err
		}
		if err := visitor.Visit(uint24(bs[15:18])); err != nil {
			return err
		}
		if err := visitor.Visit(uint24(bs[18:21])); err != nil {
			return err
		}
		if err := visitor.Visit(uint24(bs[21:24])); err != nil {
			return err
		}
	}

	for ; i < count; i++ {
		_, err := in.Read(bs[:3])
		if err != nil {
			return err
		}
		if err := visitor.Visit(uint24(bs[:3])); err != nil {
			return err
		}
	}

	return nil
}
