package main

import (
	"fmt"
	"github.com/geange/lucene-go/core/util/fst"
)

type T *fst.Pair[*fst.Pair[int64, int64], *fst.Pair[int64, int64]]

func main() {
	// *fst.Fst[*fst.Pair[*fst.Pair[int64, int64], *fst.Pair[int64, int64]]]
	posIntOutputs := fst.NewPositiveIntOutputs()
	outputsOuter := fst.NewPairOutputs[int64, int64](posIntOutputs, posIntOutputs)
	outputsInner := fst.NewPairOutputs[int64, int64](posIntOutputs, posIntOutputs)
	outputs := fst.NewPairOutputs[*fst.Pair[int64, int64], *fst.Pair[int64, int64]](outputsOuter, outputsInner)

	builder := fst.NewBuilder[*fst.Pair[*fst.Pair[int64, int64], *fst.Pair[int64, int64]]](fst.BYTE1, outputs)

	err := builder.Add([]rune("123456789"), fst.NewPair(fst.NewPair(int64(0), int64(1)), fst.NewPair(int64(1), int64(3))))
	if err != nil {
		return
	}

	err = builder.Add([]rune("223456789"), fst.NewPair(fst.NewPair(int64(1), int64(2)), fst.NewPair(int64(2), int64(4))))
	if err != nil {
		return
	}

	err = builder.Add([]rune("223456786"), fst.NewPair(fst.NewPair(int64(2), int64(3)), fst.NewPair(int64(3), int64(5))))
	if err != nil {
		return
	}

	fstInstance, err := builder.Finish()
	if err != nil {
		panic(err)
	}

	fstEnum := fst.NewBytesRefFSTEnum(fstInstance)
	output, err := fstEnum.SeekExact([]byte("223456789"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v | %+v", output.Output.Output1, output.Output.Output2)
}
