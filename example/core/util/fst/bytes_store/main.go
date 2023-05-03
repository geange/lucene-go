package main

import (
	"fmt"

	"github.com/geange/lucene-go/core/util/fst"
)

func main() {
	bytes := fst.NewByteStore(10)

	_, err := bytes.Write([]byte("1111111112"))
	if err != nil {
		panic(err)
	}

	_, err = bytes.Write([]byte("Aaaaaaaaab"))
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
