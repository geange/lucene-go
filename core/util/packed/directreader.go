package packed

import (
	"errors"

	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
)

// DirectReader
// Retrieves an instance previously written by DirectWriter
type DirectReader struct {
}

func NewDirectReader() *DirectReader {
	return &DirectReader{}
}

func DirectReaderGetInstance(slice store.RandomAccessInput, bitsPerValue int, offset int) (types.LongValues, error) {
	switch bitsPerValue {
	case 1:
		return NewDirectPackedReader1(slice, offset), nil
	case 2:
		return NewDirectPackedReader2(slice, offset), nil
	case 4:
		return NewDirectPackedReader4(slice, offset), nil
	case 8:
		return NewDirectPackedReader8(slice, offset), nil
	case 12:
		return NewDirectPackedReader12(slice, offset), nil
	case 16:
		return NewDirectPackedReader16(slice, offset), nil
	case 20:
		return NewDirectPackedReader20(slice, offset), nil
	case 24:
		return NewDirectPackedReader24(slice, offset), nil
	case 28:
		return NewDirectPackedReader28(slice, offset), nil
	case 32:
		return NewDirectPackedReader32(slice, offset), nil
	case 40:
		return NewDirectPackedReader40(slice, offset), nil
	case 48:
		return NewDirectPackedReader48(slice, offset), nil
	case 56:
		return NewDirectPackedReader56(slice, offset), nil
	case 64:
		return NewDirectPackedReader64(slice, offset), nil
	default:
		return nil, errors.New("unsupported bitsPerValue")
	}
}

func (d *DirectReader) GetInstance(slice store.RandomAccessInput, bitsPerValue int, offset int) (LongValuesReader, error) {
	switch bitsPerValue {
	case 1:
		return NewDirectPackedReader1(slice, offset), nil
	case 2:
		return NewDirectPackedReader2(slice, offset), nil
	case 4:
		return NewDirectPackedReader4(slice, offset), nil
	case 8:
		return NewDirectPackedReader8(slice, offset), nil
	case 12:
		return NewDirectPackedReader12(slice, offset), nil
	case 16:
		return NewDirectPackedReader16(slice, offset), nil
	case 20:
		return NewDirectPackedReader20(slice, offset), nil
	case 24:
		return NewDirectPackedReader24(slice, offset), nil
	case 28:
		return NewDirectPackedReader28(slice, offset), nil
	case 32:
		return NewDirectPackedReader32(slice, offset), nil
	case 40:
		return NewDirectPackedReader40(slice, offset), nil
	case 48:
		return NewDirectPackedReader48(slice, offset), nil
	case 56:
		return NewDirectPackedReader56(slice, offset), nil
	case 64:
		return NewDirectPackedReader64(slice, offset), nil
	default:
		return nil, errors.New("unsupported bitsPerValue")
	}
}

type LongValuesReader interface {
	Get(index int) (uint64, error)
}

var _ LongValuesReader = &DirectPackedReader1{}

type DirectPackedReader1 struct {
	in     store.RandomAccessInput
	offset int
}

func NewDirectPackedReader1(in store.RandomAccessInput, offset int) *DirectPackedReader1 {
	return &DirectPackedReader1{in: in, offset: offset}
}

func (d *DirectPackedReader1) Get(index int) (uint64, error) {
	shift := 7 - (index & 7)

	b, err := d.in.ReadU8(int64(d.offset + (index >> 3)))
	if err != nil {
		return 0, err
	}

	return uint64(b>>shift) & 0x1, nil
}

var _ LongValuesReader = &DirectPackedReader2{}

type DirectPackedReader2 struct {
	in     store.RandomAccessInput
	offset int
}

func NewDirectPackedReader2(in store.RandomAccessInput, offset int) *DirectPackedReader2 {
	return &DirectPackedReader2{in: in, offset: offset}
}

func (d *DirectPackedReader2) Get(index int) (uint64, error) {
	//  int shift = (3 - (int)(index & 3)) << 1;
	//        return (in.readByte(offset + (index >>> 2)) >>> shift) & 0x3;
	shift := (3 - (index & 3)) << 1

	b, err := d.in.ReadU8(int64(d.offset + (index >> 2)))
	if err != nil {
		return 0, err
	}

	return uint64(b>>shift) & 0x3, nil
}

var _ LongValuesReader = &DirectPackedReader4{}

type DirectPackedReader4 struct {
	in     store.RandomAccessInput
	offset int
}

func NewDirectPackedReader4(in store.RandomAccessInput, offset int) *DirectPackedReader4 {
	return &DirectPackedReader4{in: in, offset: offset}
}

func (d *DirectPackedReader4) Get(index int) (uint64, error) {
	shift := ((index + 1) & 1) << 2

	b, err := d.in.ReadU8(int64(d.offset + (index >> 1)))
	if err != nil {
		return 0, err
	}
	return uint64(b>>shift) & 0xF, nil
}

var _ LongValuesReader = &DirectPackedReader8{}

type DirectPackedReader8 struct {
	in     store.RandomAccessInput
	offset int
}

func NewDirectPackedReader8(in store.RandomAccessInput, offset int) *DirectPackedReader8 {
	return &DirectPackedReader8{in: in, offset: offset}
}

func (d *DirectPackedReader8) Get(index int) (uint64, error) {
	b, err := d.in.ReadU8(int64(d.offset + index))
	if err != nil {
		return 0, err
	}
	return uint64(b), nil
}

var _ LongValuesReader = &DirectPackedReader12{}

type DirectPackedReader12 struct {
	in     store.RandomAccessInput
	offset int
}

func NewDirectPackedReader12(in store.RandomAccessInput, offset int) *DirectPackedReader12 {
	return &DirectPackedReader12{in: in, offset: offset}
}

func (d *DirectPackedReader12) Get(index int) (uint64, error) {
	offset := (index * 12) >> 3
	shift := ((index + 1) & 1) << 2

	b, err := d.in.ReadU16(int64(d.offset + offset))
	if err != nil {
		return 0, err
	}
	return uint64(b>>shift) & 0xFFF, nil
}

var _ LongValuesReader = &DirectPackedReader16{}

type DirectPackedReader16 struct {
	in     store.RandomAccessInput
	offset int
}

func NewDirectPackedReader16(in store.RandomAccessInput, offset int) *DirectPackedReader16 {
	return &DirectPackedReader16{in: in, offset: offset}
}

func (d *DirectPackedReader16) Get(index int) (uint64, error) {
	u16, err := d.in.ReadU16(int64(d.offset + (index << 1)))
	if err != nil {
		return 0, err
	}
	return uint64(u16), nil
}

var _ LongValuesReader = &DirectPackedReader20{}

type DirectPackedReader20 struct {
	in     store.RandomAccessInput
	offset int
}

func NewDirectPackedReader20(in store.RandomAccessInput, offset int) *DirectPackedReader20 {
	return &DirectPackedReader20{in: in, offset: offset}
}

func (d *DirectPackedReader20) Get(index int) (uint64, error) {
	offset := (index * 20) >> 3
	u32, err := d.in.ReadU32(int64(d.offset + offset))
	if err != nil {
		return 0, err
	}
	v := u32 >> 8
	shift := ((index + 1) & 1) << 2
	return uint64(v>>shift) & 0xFFFFF, nil
}

var _ LongValuesReader = &DirectPackedReader24{}

type DirectPackedReader24 struct {
	in     store.RandomAccessInput
	offset int
}

func NewDirectPackedReader24(in store.RandomAccessInput, offset int) *DirectPackedReader24 {
	return &DirectPackedReader24{in: in, offset: offset}
}

func (d *DirectPackedReader24) Get(index int) (uint64, error) {
	u32, err := d.in.ReadU32(int64(d.offset + index*3))
	if err != nil {
		return 0, err
	}
	return uint64(u32) >> 8, nil
}

var _ LongValuesReader = &DirectPackedReader28{}

type DirectPackedReader28 struct {
	in     store.RandomAccessInput
	offset int
}

func NewDirectPackedReader28(in store.RandomAccessInput, offset int) *DirectPackedReader28 {
	return &DirectPackedReader28{in: in, offset: offset}
}

func (d *DirectPackedReader28) Get(index int) (uint64, error) {

	offset := (index * 28) >> 3
	shift := (int)((index+1)&1) << 2

	u32, err := d.in.ReadU32(int64(d.offset + offset))
	if err != nil {
		return 0, err
	}
	return uint64(u32) >> shift & 0xFFFFFFF, nil
}

var _ LongValuesReader = &DirectPackedReader32{}

type DirectPackedReader32 struct {
	in     store.RandomAccessInput
	offset int
}

func NewDirectPackedReader32(in store.RandomAccessInput, offset int) *DirectPackedReader32 {
	return &DirectPackedReader32{in: in, offset: offset}
}

func (d *DirectPackedReader32) Get(index int) (uint64, error) {
	v, err := d.in.ReadU32(int64(d.offset + index<<2))
	if err != nil {
		return 0, err
	}
	return uint64(v) & 0xFFFFFFFF, nil
}

var _ LongValuesReader = &DirectPackedReader40{}

type DirectPackedReader40 struct {
	in     store.RandomAccessInput
	offset int
}

func NewDirectPackedReader40(in store.RandomAccessInput, offset int) *DirectPackedReader40 {
	return &DirectPackedReader40{in: in, offset: offset}
}

func (d *DirectPackedReader40) Get(index int) (uint64, error) {
	v, err := d.in.ReadU64(int64(d.offset + index*5))
	if err != nil {
		return 0, err
	}
	return v >> 24, nil
}

var _ LongValuesReader = &DirectPackedReader48{}

type DirectPackedReader48 struct {
	in     store.RandomAccessInput
	offset int
}

func NewDirectPackedReader48(in store.RandomAccessInput, offset int) *DirectPackedReader48 {
	return &DirectPackedReader48{in: in, offset: offset}
}

func (d *DirectPackedReader48) Get(index int) (uint64, error) {
	// return in.readLong(this.offset + index * 6) >>> 16;
	v, err := d.in.ReadU64(int64(d.offset + index*6))
	if err != nil {
		return 0, err
	}
	return v >> 16, nil
}

var _ LongValuesReader = &DirectPackedReader56{}

type DirectPackedReader56 struct {
	in     store.RandomAccessInput
	offset int
}

func NewDirectPackedReader56(in store.RandomAccessInput, offset int) *DirectPackedReader56 {
	return &DirectPackedReader56{in: in, offset: offset}
}

func (d *DirectPackedReader56) Get(index int) (uint64, error) {
	v, err := d.in.ReadU64(int64(d.offset + index*7))
	if err != nil {
		return 0, err
	}
	return v >> 8, nil
}

var _ LongValuesReader = &DirectPackedReader64{}

type DirectPackedReader64 struct {
	in     store.RandomAccessInput
	offset int
}

func NewDirectPackedReader64(in store.RandomAccessInput, offset int) *DirectPackedReader64 {
	return &DirectPackedReader64{in: in, offset: offset}
}

func (d *DirectPackedReader64) Get(index int) (uint64, error) {
	return d.in.ReadU64(int64(d.offset + index<<3))
}
