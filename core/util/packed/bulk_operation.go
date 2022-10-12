package packed

import (
	"errors"
	"math"
)

type BulkOperation interface {
	Decoder
	Encoder
}

var (
	packedBulkOps = []BulkOperation{
		NewBulkOperationPacked1(),
		NewBulkOperationPacked2(),
		NewBulkOperationPacked3(),
		NewBulkOperationPacked4(),
		NewBulkOperationPacked5(),
		NewBulkOperationPacked6(),
		NewBulkOperationPacked7(),
		NewBulkOperationPacked8(),
		NewBulkOperationPacked9(),
		NewBulkOperationPacked10(),
		NewBulkOperationPacked11(),
		NewBulkOperationPacked12(),
		NewBulkOperationPacked13(),
		NewBulkOperationPacked14(),
		NewBulkOperationPacked15(),
		NewBulkOperationPacked16(),
		NewBulkOperationPacked17(),
		NewBulkOperationPacked18(),
		NewBulkOperationPacked19(),
		NewBulkOperationPacked20(),
		NewBulkOperationPacked21(),
		NewBulkOperationPacked22(),
		NewBulkOperationPacked23(),
		NewBulkOperationPacked24(),
		NewBulkOperationPacked(25),
		NewBulkOperationPacked(26),
		NewBulkOperationPacked(27),
		NewBulkOperationPacked(28),
		NewBulkOperationPacked(29),
		NewBulkOperationPacked(30),
		NewBulkOperationPacked(31),
		NewBulkOperationPacked(32),
		NewBulkOperationPacked(33),
		NewBulkOperationPacked(34),
		NewBulkOperationPacked(35),
		NewBulkOperationPacked(36),
		NewBulkOperationPacked(37),
		NewBulkOperationPacked(38),
		NewBulkOperationPacked(39),
		NewBulkOperationPacked(40),
		NewBulkOperationPacked(41),
		NewBulkOperationPacked(42),
		NewBulkOperationPacked(43),
		NewBulkOperationPacked(44),
		NewBulkOperationPacked(45),
		NewBulkOperationPacked(46),
		NewBulkOperationPacked(47),
		NewBulkOperationPacked(48),
		NewBulkOperationPacked(49),
		NewBulkOperationPacked(50),
		NewBulkOperationPacked(51),
		NewBulkOperationPacked(52),
		NewBulkOperationPacked(53),
		NewBulkOperationPacked(54),
		NewBulkOperationPacked(55),
		NewBulkOperationPacked(56),
		NewBulkOperationPacked(57),
		NewBulkOperationPacked(58),
		NewBulkOperationPacked(59),
		NewBulkOperationPacked(60),
		NewBulkOperationPacked(61),
		NewBulkOperationPacked(62),
		NewBulkOperationPacked(63),
		NewBulkOperationPacked(64),
	}

	packedSingleBlockBulkOps = []BulkOperation{
		NewBulkOperationPackedSingleBlock(1),
		NewBulkOperationPackedSingleBlock(2),
		NewBulkOperationPackedSingleBlock(3),
		NewBulkOperationPackedSingleBlock(4),
		NewBulkOperationPackedSingleBlock(5),
		NewBulkOperationPackedSingleBlock(6),
		NewBulkOperationPackedSingleBlock(7),
		NewBulkOperationPackedSingleBlock(8),
		NewBulkOperationPackedSingleBlock(9),
		NewBulkOperationPackedSingleBlock(10),
		nil,
		NewBulkOperationPackedSingleBlock(12),
		nil,
		nil,
		nil,
		NewBulkOperationPackedSingleBlock(16),
		nil,
		nil,
		nil,
		nil,
		NewBulkOperationPackedSingleBlock(21),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		NewBulkOperationPackedSingleBlock(32),
	}
)

func Of(format Format, bitsPerValue int) (BulkOperation, error) {
	switch format.(type) {
	case *formatPacked:
		if packedBulkOps[bitsPerValue-1] != nil {
			return packedBulkOps[bitsPerValue-1], nil
		}
	case *formatPackedSingleBlock:
		if packedSingleBlockBulkOps[bitsPerValue-1] != nil {
			return packedSingleBlockBulkOps[bitsPerValue-1], nil
		}
	}
	return nil, errors.New("AssertionError")
}

type bulkOperation struct {
	decoder Decoder
}

func writeLong(block uint64, blocks []byte, blocksOffset int) int {
	for j := 1; j <= 8; j++ {
		blocks[blocksOffset] = byte(block >> (64 - (j << 3)))
		blocksOffset++
	}
	return blocksOffset
}

// For every number of bits per value, there is a minimum number of blocks (b) / values (v) you need to write in order to reach the next block boundary: - 16 bits per value -> b=2, v=1 - 24 bits per value -> b=3, v=1 - 50 bits per value -> b=25, v=4 - 63 bits per value -> b=63, v=8 - ... A bulk read consists in copying iterations*v values that are contained in iterations*b blocks into a long[] (higher values of iterations are likely to yield a better throughput): this requires n * (b + 8v) bytes of memory. This method computes iterations as ramBudget / (b + 8v) (since a long is 8 bytes).
func (b *bulkOperation) computeIterations(valueCount, ramBudget int) int {
	iterations := ramBudget / (b.decoder.ByteBlockCount() + 8*b.decoder.ByteValueCount())
	if iterations == 0 {
		// at least 1
		return 1
	}

	if (iterations-1)*b.decoder.ByteValueCount() >= valueCount {
		// don't allocate for more than the size of the reader
		return int(math.Ceil(float64(valueCount) / float64(b.decoder.ByteValueCount())))
	}

	return iterations
}
