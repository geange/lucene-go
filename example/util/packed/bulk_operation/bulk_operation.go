package main

import (
	"fmt"
	"github.com/geange/lucene-go/core/util/packed"
)

func main() {
	op := packed.NewBulkOperationPacked(8)

	blocks := make([]uint64, 8)
	values := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	fmt.Println(values)

	op.EncodeLongToLong(values, blocks, 2)

	decodeValues := make([]uint64, 16)
	op.DecodeLongToLong(blocks, decodeValues, 2)
	fmt.Println(decodeValues)
}
