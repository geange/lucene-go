package index

type ParallelPostingsArray struct {
	size          int
	textStarts    []int
	addressOffset []int
	byteStarts    []int
}
