package index

const BYTES_PER_POSTING = 3 * 4

type ParallelPostingsArray interface {
	BytesPerPosting() int
	GetTextStarts(index int) int
	GetAddressOffset(index int) int
	GetByteStarts(index int) int
	SetTextStarts(termID, v int)
	SetAddressOffset(termID, v int)
	SetByteStarts(termID, v int)
	TextStarts() []uint32
	Grow()
}

type BaseParallelPostingsArray struct {
	textStarts    []uint32 // maps term ID to the terms's text start in the bytesHash
	addressOffset []int    // maps term ID to current stream address
	byteStarts    []int    // maps term ID to stream start offset in the byte pool
}

func NewBaseParallelPostingsArray() *BaseParallelPostingsArray {
	return &BaseParallelPostingsArray{
		textStarts:    []uint32{},
		addressOffset: []int{},
		byteStarts:    []int{},
	}
}

func (p *BaseParallelPostingsArray) GetTextStarts(index int) int {
	return int(p.textStarts[index])
}

func (p *BaseParallelPostingsArray) TextStarts() []uint32 {
	return p.textStarts
}

func (p *BaseParallelPostingsArray) ByteStarts() []int {
	return p.byteStarts
}

func (p *BaseParallelPostingsArray) AddressOffset() []int {
	return p.addressOffset
}

func (p *BaseParallelPostingsArray) GetAddressOffset(index int) int {
	return p.addressOffset[index]
}

func (p *BaseParallelPostingsArray) GetByteStarts(index int) int {
	return p.byteStarts[index]
}

func (p *BaseParallelPostingsArray) SetTextStarts(termID, v int) {
	if termID >= len(p.textStarts) {
		size := termID - len(p.textStarts) + 1
		p.textStarts = append(p.textStarts, make([]uint32, size)...)
	}
	p.textStarts[termID] = uint32(v)
}

func (p *BaseParallelPostingsArray) SetAddressOffset(termID, v int) {
	if termID >= len(p.addressOffset) {
		size := termID - len(p.addressOffset) + 1
		p.addressOffset = append(p.addressOffset, make([]int, size)...)
	}

	p.addressOffset[termID] = v
}

func (p *BaseParallelPostingsArray) SetByteStarts(termID, v int) {
	if termID >= len(p.byteStarts) {
		size := termID - len(p.byteStarts) + 1
		p.byteStarts = append(p.byteStarts, make([]int, size)...)
	}
	p.byteStarts[termID] = v
}

func (p *BaseParallelPostingsArray) Grow() {
	p.textStarts = append(p.textStarts, 0)
	p.byteStarts = append(p.byteStarts, 0)
	p.addressOffset = append(p.addressOffset, 0)
}
