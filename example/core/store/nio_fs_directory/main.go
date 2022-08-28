package main

import (
	"fmt"
	"github.com/geange/lucene-go/core/store"
)

func main() {

	directory, err := store.NewNIOFSDirectory("data")
	if err != nil {
		panic(err)
	}
	output, err := directory.CreateOutput("file.txt", nil)
	if err != nil {
		panic(err)
	}
	if err := output.WriteString("xxxxxxxx"); err != nil {
		panic(err)
	}
	if err := output.Close(); err != nil {
		return
	}

	input, err := directory.OpenInput("file.txt", nil)
	if err != nil {
		panic(err)
	}
	text, err := input.ReadString()
	if err != nil {
		panic(err)
	}
	fmt.Println(text)
}
