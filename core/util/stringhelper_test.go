package util

import (
	"fmt"
	"math"
	"math/big"
	"testing"
)

func TestRandomId(t *testing.T) {
	id := big.NewInt(int64(math.MaxInt64))
	fmt.Println(id.String())
	fmt.Println(len(id.Bytes()))
	id.Add(id, one)
	fmt.Println(id.String())
	fmt.Println(len(id.Bytes()))

}
