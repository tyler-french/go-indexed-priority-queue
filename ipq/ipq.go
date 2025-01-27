package ipq

import (
	"errors"
	"sync"
)

// Comparator function to define heap order (min-heap or max-heap)
type Comparator[T any] func(a, b T) bool

// IndexedPriorityQueue represents an indexed priority queue.
type IndexedPriorityQueue[T any] struct {
	mu   sync.Mutex
	keys map[int]int // Maps user index -> heap index
	heap []entry[T]
	comp Comparator[T] // Custom comparator
}

type entry[T any] struct {
	index int
	value T
}

// New creates a new IndexedPriorityQueue with the given comparator.
// for example, an indexed priority queue for integers would simply be:
// `ipq.New(func(a, b int) bool { return a < b })`
//
// There are also default constructors for common types: int, string, and float
func New[T any](comp Comparator[T]) (*IndexedPriorityQueue[T], error) {
	if comp == nil {
		return nil, errors.New("comparator can't be empty")
	}
	return &IndexedPriorityQueue[T]{
		keys: make(map[int]int),
		comp: comp,
	}, nil
}

// NewInt64 creates a new priority queue for integers.
func NewInt64() *IndexedPriorityQueue[int64] {
	ipq, err := New(func(a, b int64) bool { return a < b })
	if err != nil {
		panic("not reachable")
	}
	return ipq
}

// NewInt creates a new priority queue for integers.
func NewInt() *IndexedPriorityQueue[int] {
	ipq, err := New(func(a, b int) bool { return a < b })
	if err != nil {
		panic("not reachable")
	}
	return ipq
}

// NewString creates a new queue for strings.
func NewString() *IndexedPriorityQueue[string] {
	ipq, err := New(func(a, b string) bool { return a < b })
	if err != nil {
		panic("not reachable")
	}
	return ipq
}

// NewFloat64 creates a new queue for float64s
func NewFloat64() *IndexedPriorityQueue[float64] {
	ipq, err := New(func(a, b float64) bool { return a < b })
	if err != nil {
		panic("not reachable")
	}
	return ipq
}

// Push inserts a new element or updates an existing one.
func (pq *IndexedPriorityQueue[T]) Push(index int, value T) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if i, exists := pq.keys[index]; exists {
		pq.heap[i].value = value
		pq.up(i)
		pq.down(i)
		return
	}

	pq.keys[index] = len(pq.heap)
	pq.heap = append(pq.heap, entry[T]{index, value})
	pq.up(len(pq.heap) - 1)
}

// Pop removes and returns the top element (min or max).
func (pq *IndexedPriorityQueue[T]) Pop() (int, T, error) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if pq.Empty() {
		var zero T
		return -1, zero, errors.New("priority queue is empty")
	}

	top := pq.heap[0]
	pq.swap(0, len(pq.heap)-1)
	pq.heap = pq.heap[:len(pq.heap)-1]
	delete(pq.keys, top.index)
	if len(pq.heap) > 0 {
		pq.down(0)
	}
	return top.index, top.value, nil
}

// Get retrieves the value associated with the given index.
func (pq *IndexedPriorityQueue[T]) Get(index int) (T, error) {
	i, exists := pq.keys[index]
	if !exists {
		var zero T
		return zero, errors.New("index not found")
	}

	return pq.heap[i].value, nil
}

// DecreaseKey updates an element’s value (if lower priority).
func (pq *IndexedPriorityQueue[T]) DecreaseKey(index int, newValue T) error {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	i, exists := pq.keys[index]
	if !exists {
		return errors.New("index not found")
	}
	if pq.comp(pq.heap[i].value, newValue) { // Already lower priority
		return nil
	}

	pq.heap[i].value = newValue
	pq.up(i)
	return nil
}

// IncreaseKey updates an element’s value (if higher priority).
func (pq *IndexedPriorityQueue[T]) IncreaseKey(index int, newValue T) error {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	i, exists := pq.keys[index]
	if !exists {
		return errors.New("index not found")
	}
	if pq.comp(newValue, pq.heap[i].value) { // Already higher priority
		return nil
	}

	pq.heap[i].value = newValue
	pq.down(i)
	return nil
}

// Delete removes an element by index.
func (pq *IndexedPriorityQueue[T]) Delete(index int) error {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	i, exists := pq.keys[index]
	if !exists {
		return errors.New("index not found")
	}

	pq.swap(i, len(pq.heap)-1)
	pq.heap = pq.heap[:len(pq.heap)-1]
	delete(pq.keys, index)
	if i < len(pq.heap) {
		pq.up(i)
		pq.down(i)
	}
	return nil
}

// Peek returns the top element without removing it.
func (pq *IndexedPriorityQueue[T]) Peek() (int, T, error) {
	if pq.Empty() {
		var zero T
		return -1, zero, errors.New("priority queue is empty")
	}
	return pq.heap[0].index, pq.heap[0].value, nil
}

func (pq *IndexedPriorityQueue[T]) Len() int {
	return len(pq.heap)
}

func (pq *IndexedPriorityQueue[T]) Empty() bool {
	return pq.Len() == 0
}

func (pq *IndexedPriorityQueue[T]) up(i int) {
	for i > 0 {
		p := (i - 1) / 2
		if pq.comp(pq.heap[p].value, pq.heap[i].value) {
			break
		}
		pq.swap(i, p)
		i = p
	}
}

// down restores the heap property by "sifting down" the element at index i.
// It compares the element at i with its children, and if one of the children
// has a higher priority (smaller value for a min-heap), it swaps the elements
// and continues this process recursively down the heap until the heap property is restored.
// 
// This function is typically used after removing or modifying an element at the root of the heap,
// to ensure the heap property is maintained throughout the structure.
//
// Complexity: O(log n), where n is the number of elements in the heap.
func (pq *IndexedPriorityQueue[T]) down(i int) {
    n := len(pq.heap) // Get the total number of elements in the heap
    for {
        // Calculate indices for the left (l) and right (r) children of the current node (i)
        l, r, smallest := 2*i+1, 2*i+2, i

        // Check if the left child exists and has a higher priority (smaller value) than the current node
        if l < n && pq.comp(pq.heap[l].value, pq.heap[smallest].value) {
            smallest = l // Update smallest if left child is higher priority
        }

        // Check if the right child exists and has a higher priority than the smallest value so far
        if r < n && pq.comp(pq.heap[r].value, pq.heap[smallest].value) {
            smallest = r // Update smallest if right child is higher priority
        }

        // If the current node is the smallest, the heap property is satisfied, so exit the loop
        if smallest == i {
            break
        }

        // Swap the current node with the smallest child
        pq.swap(i, smallest)

        // Move to the next level in the heap (i.e., the index of the swapped child)
        i = smallest
    }
}


func (pq *IndexedPriorityQueue[T]) swap(i, j int) {
	pq.keys[pq.heap[i].index], pq.keys[pq.heap[j].index] = j, i
	pq.heap[i], pq.heap[j] = pq.heap[j], pq.heap[i]
}
