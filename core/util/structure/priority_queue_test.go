package structure

import (
	"fmt"
	"math"
	"testing"
)

func TestNewPriorityQueue(t *testing.T) {
	queue := NewPriorityQueue[*Struct](5, func(a, b *Struct) bool {
		return (a.A < b.A) || (a.A == b.A && a.B < b.B)
	})

	tmpTop := &Struct{1, 3}
	queue.Add(tmpTop)

	queue.Add(&Struct{
		A: 1,
		B: 2,
	})

	tmpTop.A = math.Inf(-1)

	queue.UpdateTop()
	fmt.Println(queue.Top())
}

type Struct struct {
	A float64
	B int
}
