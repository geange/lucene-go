package packed

// Decoder A decoder for packed integers.
type Decoder interface {

	// LongBlockCount The minimum number of long blocks to encode in a single iteration, when using long encoding.
	LongBlockCount() int

	// LongValueCount The number of values that can be stored in longBlockCount() long blocks.
	LongValueCount() int

	// ByteBlockCount The minimum number of byte blocks to encode in a single iteration, when using byte encoding.
	ByteBlockCount() int

	// ByteValueCount The number of values that can be stored in byteBlockCount() byte blocks.
	ByteValueCount() int

	// DecodeLongToLong Read iterations * blockCount() blocks from blocks, decode them and write iterations * valueCount() values into values.
	// Params: 	blocks – the long blocks that hold packed integer values
	//			blocksOffset – the offset where to start reading blocks
	//			values – the values buffer
	//			valuesOffset – the offset where to start writing values
	//			iterations – controls how much data to decode
	DecodeLongToLong(blocks, values []int64, iterations int)

	// DecodeByteToLong Read 8 * iterations * blockCount() blocks from blocks, decode them and write iterations * valueCount() values into values.
	// Params: 	blocks – the long blocks that hold packed integer values
	//			blocksOffset – the offset where to start reading blocks
	//			values – the values buffer
	//			valuesOffset – the offset where to start writing values
	//			iterations – controls how much data to decode
	DecodeByteToLong(blocks []byte, values []int64, iterations int)

	// DecodeLongToInt Read iterations * blockCount() blocks from blocks, decode them and write iterations * valueCount() values into values.
	// Params: 	blocks – the long blocks that hold packed integer values blocksOffset – the offset where to start reading blocks values – the values buffer valuesOffset – the offset where to start writing values iterations – controls how much data to decode
	//DecodeLongToInt(blocks []int64, values []int, iterations int)

	// DecodeByteToInt Read 8 * iterations * blockCount() blocks from blocks, decode them and write iterations * valueCount() values into values.
	// Params: 	blocks – the long blocks that hold packed integer values
	//			blocksOffset – the offset where to start reading blocks
	//			values – the values buffer
	//			valuesOffset – the offset where to start writing values
	//			iterations – controls how much data to decode
	DecodeByteToInt(blocks []byte, values []int32, iterations int)
}
