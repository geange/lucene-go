package index

// Codec Encodes/decodes an inverted index segment.
// Note, when extending this class, the name (getName) is written into the index. In order for the segment to be read, the name must resolve to your implementation via forName(String). This method uses Java's Service Provider Interface (SPI) to resolve codec names.
// If you implement your own codec, make sure that it has a no-arg constructor so SPI can load it.
// See Also: ServiceLoader
type Codec interface {
}
