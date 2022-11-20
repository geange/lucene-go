package main

import "github.com/geange/lucene-go/core/util/fst"

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

	_, err = builder.Finish()
	if err != nil {
		panic(err)
	}

}
