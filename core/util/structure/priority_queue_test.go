package structure

import (
	"fmt"
	"testing"
)

func TestNewPriorityQueue(t *testing.T) {
	queue := NewPriorityQueue[int](5, func(a, b int) bool {
		return a < b
	})
	queue.Add(3)
	queue.Add(5)
	queue.Add(0)
	queue.Add(-2)
	fmt.Println(queue.Top())
}
