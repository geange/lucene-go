package packed

// BlockPackedWriter
// A writer for large sequences of longs.
// The sequence is divided into fixed-size blocks and for each block, the difference between each value
// and the minimum value of the block is encoded using as few bits as possible. Memory usage of this
// class is proportional to the block size. Each block has an overhead between 1 and 10 bytes to store
// the minimum value and the number of bits per value of the block.
// Format:
// <BLock>BlockCount
// BlockCount: ⌈ ValueCount / BlockSize ⌉
// Block: <Header, (Ints)>
// Header: <Token, (MinValue)>
// Token: a byte, first 7 bits are the number of bits per value (bitsPerValue). If the 8th bit is 1, then MinValue (see next) is 0, otherwise MinValue and needs to be decoded
// MinValue: a zigzag-encoded  variable-length long whose value should be added to every int from the block to restore the original values
// Ints: If the number of bits per value is 0, then there is nothing to decode and all ints are equal to MinValue. Otherwise: BlockSize packed ints encoded on exactly bitsPerValue bits per value. They are the subtraction of the original values and MinValue
// 请参阅: BlockPackedReaderIterator, BlockPackedReader
// lucene.internal
type BlockPackedWriter struct {
}
