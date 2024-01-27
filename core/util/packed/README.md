# packed

Packed integer arrays and streams.

The packed package provides

- sequential and random access capable arrays of positive longs,
- routines for efficient serialization and deserialization of streams of packed integers.

The implementations provide different trade-offs between memory usage and access speed. The standard usage scenario is
replacing large int or long arrays in order to reduce the memory footprint.

## In-memory structures

### Mutable

- Only supports positive longs.
- Requires the number of bits per value to be known in advance.
- Random-access for both writing and reading.

### GrowableWriter

- Same as PackedInts.Mutable but grows the number of bits per values when needed.
- Useful to build a PackedInts.Mutable from a read-once stream of longs.

### PagedGrowableWriter

- Slices data into fixed-size blocks stored in GrowableWriters.
- Supports more than 2B values.
- You should use PackedLongValues instead if you don't need random write access.

### PackedLongValues.deltaPackedBuilder

- Can store any sequence of longs.
- Compression is good when values are close to each other.
- Supports random reads, but only sequential writes.
- Can address up to 2^42 values.

### PackedLongValues.packedBuilder

- Same as deltaPackedBuilder but assumes values are 0-based.

### PackedLongValues.monotonicBuilder

- Same as deltaPackedBuilder except that compression is good when the stream is a succession of affine functions.

## Disk-based structures

### Writer/Reader/ReaderIterator

- Only supports positive longs.
- Requires the number of bits per value to be known in advance.
- Splits the stream into fixed-size blocks.
- Supports both fast sequential access with low memory footprint with ReaderIterator and random-access by either loading
  values in memory or leaving them on disk with Reader.

### BlockPackedWriter/BlockPackedReader/BlockPackedReaderIterator

- Splits the stream into fixed-size blocks.
- Compression is good when values are close to each other.
- Can address up to 2B - blockSize values.

### MonotonicBlockPackedWriter/MonotonicBlockPackedReader

- Same as the non-monotonic variants except that compression is good when the stream is a succession of affine
  functions.
- The reason why there is no sequential access is that if you need sequential access, you should rather delta-encode and
  use BlockPackedWriter.

### PackedDataOutput/PackedDataInput

- Writes sequences of longs where each long can use any number of bits.