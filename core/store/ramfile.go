package store

// RAMFile Represents a file in RAM as a list of byte[] buffers.
// This class uses inefficient synchronization and is discouraged in favor of MMapDirectory.
// It will be removed in future versions of Lucene.
// lucene.internal
type RAMFile struct {
	buffers [][]byte
	length  int64
}

func NewRAMFile() *RAMFile {
	return &RAMFile{
		buffers: make([][]byte, 0),
		length:  0,
	}
}

func (r *RAMFile) GetLength() int64 {
	return r.length
}

func (r *RAMFile) setLength(length int64) {
	r.length = length
}

func (r *RAMFile) addBuffer(size int) []byte {
	buffer := make([]byte, size)
	r.buffers = append(r.buffers, buffer)
	return buffer
}

func (r *RAMFile) getBuffer(index int) []byte {
	if index < len(r.buffers) {
		return r.buffers[index]
	}
	return nil
}

// Expert: allocate a new buffer. Subclasses can allocate differently.
// Params: size – size of allocated buffer.
// Returns: allocated buffer.
func (r *RAMFile) numBuffers() int {
	return len(r.buffers)
}
