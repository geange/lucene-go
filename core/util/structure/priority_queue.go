package structure

import (
	"io"
	"reflect"
)

type PriorityQueue[T any] struct {
	size     int
	maxSize  int
	heap     []T
	none     T
	lessThan func(a, b T) bool
}

func NewPriorityQueue[T any](maxSize int, lessThan func(a, b T) bool) *PriorityQueue[T] {
	var a T
	return NewPriorityQueueV1(maxSize, func() T {
		return a
	}, lessThan)
}

func NewPriorityQueueV1[T any](maxSize int, supplier func() T, lessThan func(a, b T) bool) *PriorityQueue[T] {
	if maxSize < 2 {
		maxSize = 2
	}

	queue := &PriorityQueue[T]{
		maxSize:  maxSize,
		heap:     make([]T, maxSize+1),
		lessThan: lessThan,
	}
	for i := range queue.heap {
		queue.heap[i] = supplier()
	}
	return queue
}

// Add
// Adds an Object to a PriorityQueue in log(size) time.
// If one tries to add more objects than maxSize from initialize an ArrayIndexOutOfBoundsException is thrown.
// Returns: the new 'top' element in the queue.
func (p *PriorityQueue[T]) Add(element T) T {
	p.size++
	p.heap[p.size] = element
	p.upHeap(p.size)
	return p.heap[1]
}

// Top
// Returns the least element of the PriorityQueue in constant time.
func (p *PriorityQueue[T]) Top() T {
	return p.heap[1]
}

// Pop
// Removes and returns the least element of the PriorityQueue in log(size) time.
func (p *PriorityQueue[T]) Pop() (T, error) {
	if p.size > 0 {
		result := p.heap[1]        // save first value
		p.heap[1] = p.heap[p.size] // move last to first
		p.heap[p.size] = p.none    // permit GC of objects
		p.size--
		p.downHeap(1) // adjust heap
		return result, nil
	} else {
		return p.none, io.EOF
	}
}

// UpdateTop
// Should be called when the Object at top changes values. Still log(n) worst case,
// but it's at least twice as fast to
//
//	pq.top().change();
//	pq.updateTop();
//
// instead of
//
//	o = pq.pop();
//	o.change();
//	pq.push(o);
//
// Returns:
// the new 'top' element.
func (p *PriorityQueue[T]) UpdateTop() T {
	p.downHeap(1)
	return p.heap[1]
}

// UpdateTopByNewTop
// Replace the top of the pq with newTop and run updateTop().
func (p *PriorityQueue[T]) UpdateTopByNewTop(newTop T) T {
	p.heap[1] = newTop
	return p.UpdateTop()
}

// Size Returns the number of elements currently stored in the PriorityQueue.
func (p *PriorityQueue[T]) Size() int {
	return p.size
}

// Clear
// Removes all entries from the PriorityQueue.
func (p *PriorityQueue[T]) Clear() {
	for i := 0; i <= p.size; i++ {
		p.heap[i] = p.none
	}
	p.size = 0
}

// Remove
// Removes an existing element currently stored in the PriorityQueue.
// Cost is linear with the size of the queue.
//
// (A specialization of PriorityQueue which tracks element positions
// would provide a constant remove time but the trade-off would be
// extra cost to all additions/insertions)
func (p *PriorityQueue[T]) Remove(element T) bool {
	for i := 1; i <= p.size; i++ {
		if reflect.DeepEqual(p.heap[i], element) {
			p.heap[i] = p.heap[p.size]
			p.heap[p.size] = p.none // permit GC of objects
			p.size--
			if i <= p.size {
				if !p.upHeap(i) {
					p.downHeap(i)
				}
			}
			return true
		}
	}
	return false
}

func (p *PriorityQueue[T]) Iterator() Iterator[T] {
	return &PriorityQueueIterator[T]{
		i:             1,
		PriorityQueue: p,
	}
}

func (p *PriorityQueue[T]) upHeap(origPos int) bool {
	i := origPos
	node := p.heap[i] // save bottom node
	j := i >> 1
	for j > 0 && p.lessThan(node, p.heap[j]) {
		p.heap[i] = p.heap[j] // shift parents down
		i = j
		j = j >> 1
	}
	p.heap[i] = node // install saved node
	return i != origPos
}

func (p *PriorityQueue[T]) downHeap(i int) {
	node := p.heap[i] // save top node
	j := i << 1       // find smaller child
	k := j + 1
	if k <= p.size && p.lessThan(p.heap[k], p.heap[j]) {
		j = k
	}
	for j <= p.size && p.lessThan(p.heap[j], node) {
		p.heap[i] = p.heap[j] // shift up child
		i = j
		j = i << 1
		k = j + 1
		if k <= p.size && p.lessThan(p.heap[k], p.heap[j]) {
			j = k
		}
	}
	p.heap[i] = node // install saved node
}

type PriorityQueueIterator[T any] struct {
	i int
	*PriorityQueue[T]
}

func (p *PriorityQueueIterator[T]) HasNext() bool {
	return p.i <= p.size
}

func (p *PriorityQueueIterator[T]) Next() (T, error) {
	if !p.HasNext() {
		return p.none, io.EOF
	}
	idx := p.i
	p.i++
	return p.heap[idx], nil
}
