package main

import (
	"fmt"
	"github.com/geange/lucene-go/core/util/fst"
)

func main() {
	posIntOutputs := fst.NewPositiveIntOutputs()

	builder := fst.NewBuilder(fst.BYTE1, posIntOutputs)

	// "mop", "moth", "pop", "star", "stop", "top"
	// 100, 91, 72, 83, 54, 55
	err := builder.Add([]rune("mop"), int64(100))
	if err != nil {
		panic(err)
	}

	err = builder.Add([]rune("moth"), int64(91))
	if err != nil {
		panic(err)
	}

	err = builder.Add([]rune("pop"), int64(72))
	if err != nil {
		panic(err)
	}

	err = builder.Add([]rune("star"), int64(83))
	if err != nil {
		panic(err)
	}

	err = builder.Add([]rune("stop"), int64(54))
	if err != nil {
		panic(err)
	}

	err = builder.Add([]rune("top"), int64(55))
	if err != nil {
		panic(err)
	}

	fmap, err := builder.Finish()
	if err != nil {
		panic(err)
	}
	arc := new(fst.Arc)
	firstArc1, err := fmap.GetFirstArc(arc)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", firstArc1)

	err = fmap.SaveToFile("test.fst")
	if err != nil {
		panic(err)
	}

	newFST, err := fst.NewFSTFromFile("test.fst", &fst.PositiveIntOutputs{})
	if err != nil {
		panic(err)
	}

	reader, _ := newFST.GetBytesReader()

	arc2 := new(fst.Arc)
	arc2, err = newFST.GetFirstArc(arc2)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%c\n", arc2.Label())

	follow := new(fst.Arc)
	arc2, err = newFST.FindTargetArc(int('t'), arc2, follow, reader)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%c", arc2.Label())

	arc2, err = newFST.FindTargetArc(int('o'), arc2, follow, reader)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%c", arc2.Label())

	arc2, err = newFST.FindTargetArc(int('p'), arc2, follow, reader)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%c\n", arc2.Label())

}
