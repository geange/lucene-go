package main

import (
	"fmt"

	"github.com/geange/lucene-go/core/util/fst"
)

func main() {
	bytes := fst.NewBytesStore(2)

	err := bytes.WriteBytes([]byte("1111111111"))
	if err != nil {
		panic(err)
	}

	err = bytes.WriteBytes([]byte("Aaaaaaaaab"))
	if err != nil {
		panic(err)
	}

	err = bytes.MoveBytes(0, 10, 10)
	if err != nil {
		panic(err)
	}

	fmt.Println(bytes.GetPosition())

	err = bytes.WriteString("bbbbbbbbbb")
	if err != nil {
		panic(err)
	}
}
