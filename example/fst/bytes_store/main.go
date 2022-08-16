package main

import (
	"fmt"
	"github.com/geange/lucene-go/core/fst"
)

func main() {
	bytes := fst.NewBytesStore(15)

	err := bytes.WriteBytes([]byte("1111111111"))
	if err != nil {
		panic(err)
	}

	err = bytes.WriteBytes([]byte("Aaaaaaaaab"))
	if err != nil {
		panic(err)
	}

	bytes.CopyBytesToSelf(10, 0, 10)

	fmt.Println(bytes.GetPosition())
}
