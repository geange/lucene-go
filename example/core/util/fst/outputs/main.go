package main

import (
	"fmt"
	"github.com/geange/lucene-go/core/util/fst"
)

func main() {
	outputs := fst.NewPositiveIntOutputs()
	pairOutputs := fst.NewPairOutputs[int64, int64](outputs, outputs)

	common, err := pairOutputs.Common(fst.NewPair(int64(1), int64(2)), fst.NewPair(int64(5), int64(6)))
	if err != nil {
		return
	}

	fmt.Println(common.Output1, common.Output2)
}
