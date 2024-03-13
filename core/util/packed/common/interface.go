package common

type BulkOperation interface {
	Decoder
	Encoder
}

// Decoder A decoder for packed integers.
type Decoder interface {

	// LongBlockCount
	// The minimum number of long blocks to encode in a single iteration, when using long encoding.
	LongBlockCount() int

	// LongValueCount
	// The number of values that can be stored in longBlockCount() long blocks.
	LongValueCount() int

	// ByteBlockCount
	// The minimum number of byte blocks to encode in a single iteration, when using byte encoding.
	ByteBlockCount() int

	// ByteValueCount
	// The number of values that can be stored in byteBlockCount() byte blocks.
	ByteValueCount() int

	// DecodeUint64
	// Read iterations * blockCount() blocks from blocks,
	// decode them and write iterations * valueCount() values into values.
	// blocks: the long blocks that hold packed integer values
	// values: the values buffer
	// iterations: controls how much data to decode
	DecodeUint64(blocks []uint64, values []uint64, iterations int)

	// DecodeBytes
	// Read 8 * iterations * blockCount() blocks from blocks,
	// decode them and write iterations * valueCount() values into values.
	// blocks: the long blocks that hold packed integer values
	// values: the values buffer
	// iterations: controls how much data to decode
	DecodeBytes(blocks []byte, values []uint64, iterations int)
}

// Encoder An encoder for packed integers.
type Encoder interface {
	// LongBlockCount
	// The minimum number of long blocks to encode in a single iteration, when using long encoding.
	LongBlockCount() int

	// LongValueCount
	// The number of values that can be stored in longBlockCount() long blocks.
	LongValueCount() int

	// ByteBlockCount
	// The minimum number of byte blocks to encode in a single iteration, when using byte encoding.
	ByteBlockCount() int

	// ByteValueCount
	// The number of values that can be stored in byteBlockCount() byte blocks.
	ByteValueCount() int

	// EncodeUint64
	// Read iterations * valueCount() values from values, encode them and write
	// iterations * blockCount() blocks into blocks.
	// values: the values buffer
	// blocks: the long blocks that hold packed integer values
	// iterations: controls how much data to encode
	EncodeUint64(values []uint64, blocks []uint64, iterations int)

	// EncodeBytes
	// Read iterations * valueCount() values from values,
	// encode them and write 8 * iterations * blockCount() blocks into blocks.
	// values: the values buffer
	// blocks: the long blocks that hold packed integer values
	// iterations: controls how much data to encode
	EncodeBytes(values []uint64, blocks []byte, iterations int)
}
