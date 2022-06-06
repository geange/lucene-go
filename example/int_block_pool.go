package main

import (
	"fmt"
	"github.com/geange/lucene-go/core/util"
)

func main() {
	pool := util.NewIntBlockPool()

	writer := util.NewSliceWriter(pool)

	writer.StartNewSlice()

	for i := 0; i < 100; i++ {
		writer.WriteInt(i)
	}

	writer.StartNewSlice()

	for i := 0; i < 100; i++ {
		writer.WriteInt(i)
	}

	fmt.Println(pool)
}
