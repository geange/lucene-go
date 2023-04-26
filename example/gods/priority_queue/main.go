package main

import (
	"fmt"
	"github.com/emirpasic/gods/queues/priorityqueue"
)

func main() {
	queue := priorityqueue.NewWith(func(a, b interface{}) int {
		n1, n2 := a.(int), b.(int)
		if n1 < n2 {
			return -1
		}
		if n1 == n2 {
			return 0
		}
		return 1
	})

	queue.Enqueue(1)
	queue.Enqueue(8)
	queue.Enqueue(6)
	queue.Enqueue(0)

	value, _ := queue.Peek()
	fmt.Println(value)
}
