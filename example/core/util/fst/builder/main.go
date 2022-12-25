package main

import (
	"fmt"

	"github.com/geange/lucene-go/core/util/fst"
)

func main() {
	posIntOutputs := fst.NewPositiveIntOutputs[int64]()

	builder := fst.NewBuilder[int64](fst.BYTE1, posIntOutputs)

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
	arc := new(fst.Arc[int64])
	firstArc1, err := fmap.GetFirstArc(arc)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", firstArc1)

	err = fmap.SaveToFile("test.fst")
	if err != nil {
		panic(err)
	}

	{
		newFST, err := fst.NewFSTFromFile[int64]("test.fst", &fst.PositiveIntOutputs[int64]{})
		if err != nil {
			panic(err)
		}

		reader, _ := newFST.GetBytesReader()

		follow := new(fst.Arc[int64])
		follow, err = newFST.GetFirstArc(follow)
		if err != nil {
			panic(err)
		}

		fmt.Printf("%c\n", follow.Label())

		arc := new(fst.Arc[int64])
		follow, err = newFST.FindTargetArc(int('t'), follow, arc, reader)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%c", follow.Label())

		follow, err = newFST.FindTargetArc(int('o'), follow, arc, reader)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%c", follow.Label())

		follow, err = newFST.FindTargetArc(int('p'), follow, arc, reader)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%c\n", follow.Label())
	}

}
