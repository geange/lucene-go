package packed

// Encoder An encoder for packed integers.
type Encoder interface {

	// EncodeLongToLong Read iterations * valueCount() values from values, encode them and write
	// iterations * blockCount() blocks into blocks.
	// Params: 	values – the values buffer
	//			valuesOffset – the offset where to start reading values
	//			blocks – the long blocks that hold packed integer values
	//			blocksOffset – the offset where to start writing blocks
	//			iterations – controls how much data to encode
	EncodeLongToLong(values, blocks []int64, iterations int)

	// EncodeLongToBytes Read iterations * valueCount() values from values,
	// encode them and write 8 * iterations * blockCount() blocks into blocks.
	// Params: 	values – the values buffer
	//			valuesOffset – the offset where to start reading values
	//			blocks – the long blocks that hold packed integer values
	//			blocksOffset – the offset where to start writing blocks
	//			iterations – controls how much data to encode
	EncodeLongToBytes(values []int64, blocks []byte, iterations int)

	// EncodeIntToBytes Read iterations * valueCount() values from values, encode them and write 8 * iterations * blockCount() blocks into blocks.
	// Params: 	values – the values buffer
	//			valuesOffset – the offset where to start reading values
	//			blocks – the long blocks that hold packed integer values
	//			blocksOffset – the offset where to start writing blocks
	//			iterations – controls how much data to encode
	EncodeIntToBytes(values []int32, blocks []byte, iterations int)
}
