package ipq

import (
	"errors"
	"sync"
)

// Comparator function to define heap order (min-heap or max-heap)
type Comparator[T any] func(a, b T) bool

// IndexedPriorityQueue represents an indexed priority queue.
// It uses a binary heap structure to store the elements.
// The heap is organized such that each element has a higher priority (smaller value)
// than its children, ensuring efficient retrieval of the highest-priority element.
// Items can be retrieved by their key.
type IndexedPriorityQueue[T any] struct {
	mu   sync.RWMutex
	keys map[int]int   // Maps user index -> heap index
	heap heap[T]       // Binary heap storing the elements
	comp Comparator[T] // Custom comparator function to determine priority
}

// Heap Structure:
// The heap is stored as a slice where the elements satisfy the heap property:
// - The parent node is always smaller (has higher priority) than its children (for a min-heap).
// - The left child of a node at index i is at index 2*i + 1, and the right child is at index 2*i + 2.
//
// Example of a Min-Heap (as a tree):
//
//	      10
//	     /  \
//	   20    15
//	  /  \
//	30    40
//
// The corresponding slice representation (heap array) of the above tree is:
// heap = [10, 20, 15, 30, 40]
//
// In the heap slice, elements are stored in the order they appear level by level (left to right):
// - The root element (10) is at index 0.
// - Its children (20 and 15) are at indices 1 and 2, respectively.
// - The next level has the children of 20 and 15, i.e., 30 and 40, at indices 3 and 4.
//
// How the `keys` map works:
// The `keys` map stores the mapping of the user's index to the heap index. For example,
// if an element with user index 2 is at index 3 in the heap, `keys[2]` would be 3.
//
// Insertion Example:
// Let's consider inserting the value 25 into the heap. Here's how it would work:
//
// Initial heap (as a slice): [10, 20, 15, 30, 40]
// - The new element 25 is added to the end of the slice: [10, 20, 15, 30, 40, 25].
// - The "up" operation is then performed to restore the heap property.
//   - 25 is compared with its parent (30). Since 25 is smaller, they are swapped.
//   - New heap after swap: [10, 20, 15, 25, 40, 30].
//   - The heap property is now restored, and the insertion is complete.
//
// The resulting heap after the insertion:
//
//	      10
//	     /  \
//	   20    15
//	  /  \   /
//	25   40 30
//
// Heap as a slice after insertion:
// heap = [10, 20, 15, 25, 40, 30]
//
// The `comp` function is used to define the custom comparison logic for determining priority.
type heap[T any] []entry[T]

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

// Push inserts a new element or updates an existing one in the priority queue.
// If the element with the given index already exists, its value is updated,
// and the heap is adjusted to restore the heap property. If the element does
// not exist, a new element is appended to the heap and the heap property is restored.
// This function ensures that the heap remains a valid priority queue after each operation.
func (pq *IndexedPriorityQueue[T]) Push(index int, value T) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if i, exists := pq.keys[index]; exists {
		pq.heap[i].value = value

		// Call up() to restore the heap property by moving the updated element upwards
		pq.up(i)

		// Call down() to ensure the heap property is maintained in case the update
		// caused the element to violate the heap property in its children
		pq.down(i)
		return
	}

	// If the element does not exist, we add it as a new entry
	pq.keys[index] = len(pq.heap)
	pq.heap = append(pq.heap, entry[T]{index, value})

	// Call up() to restore the heap property by moving the newly inserted element upwards
	// This ensures that the new element is in the correct position according to its priority
	pq.up(len(pq.heap) - 1)
}

// Pop removes and returns the top element (min or max) from the priority queue.
// The function returns the index and value of the removed element, or an error if the queue is empty.
// After removing the top element, it restores the heap property to maintain the queue's validity.
func (pq *IndexedPriorityQueue[T]) Pop() (int, T, error) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if pq.Empty() {
		var zero T
		return -1, zero, errors.New("priority queue is empty")
	}

	top := pq.heap[0]

	// Swap the top element with the last element in the heap (to remove it more efficiently)
	pq.swap(0, len(pq.heap)-1)

	// Remove the last element (which was originally the top element) from the heap slice
	pq.heap = pq.heap[:len(pq.heap)-1]

	delete(pq.keys, top.index)

	// restore the heap property
	if len(pq.heap) > 0 {
		pq.down(0)
	}
	return top.index, top.value, nil
}

// Get retrieves the value associated with the given index.
func (pq *IndexedPriorityQueue[T]) Get(index int) (T, error) {
	pq.mu.RLock()
	defer pq.mu.RUnlock()

	i, exists := pq.keys[index]
	if !exists {
		var zero T
		return zero, errors.New("index not found")
	}

	return pq.heap[i].value, nil
}

// DecreaseKey updates an element’s value if the new value represents a lower priority.
// This is typically used to decrease the priority of an element (e.g., for a min-heap).
// After updating the value, it ensures the heap property is maintained by moving the element upwards if necessary.
func (pq *IndexedPriorityQueue[T]) DecreaseKey(index int, newValue T) error {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	i, exists := pq.keys[index]
	if !exists {
		return errors.New("index not found")
	}

	// new value is already lower priority
	if pq.comp(pq.heap[i].value, newValue) {
		return nil
	}

	// Update the element's value with the new lower priority value
	pq.heap[i].value = newValue

	// Call up() to restore the heap property by moving the updated element upwards
	pq.up(i)

	return nil
}

// IncreaseKey updates an element’s value if the new value represents a higher priority.
// This is typically used to increase the priority of an element (e.g., for a min-heap).
// After updating the value, it ensures the heap property is maintained by moving the element downwards if necessary.
func (pq *IndexedPriorityQueue[T]) IncreaseKey(index int, newValue T) error {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	i, exists := pq.keys[index]
	if !exists {
		return errors.New("index not found")
	}

	// new value is already higher priority
	if pq.comp(newValue, pq.heap[i].value) {
		return nil
	}

	// Update the element's value with the new higher priority value
	pq.heap[i].value = newValue

	// Call down() to restore the heap property by moving the updated element downwards
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
	pq.mu.RLock()
	defer pq.mu.RUnlock()

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

// up restores the heap property by "sifting up" the element at index i.
// It compares the element at i with its parent, and if the parent has a lower priority
// (greater value for a min-heap), it swaps the two elements and continues the process upwards
// until the heap property is restored or the root of the heap is reached.
func (pq *IndexedPriorityQueue[T]) up(i int) {
	// Continue moving up as long as the current node is not the root (i > 0)
	for i > 0 {
		// Calculate the index of the parent node
		parent := (i - 1) / 2

		// If the parent has a higher priority (smaller value), the heap property is satisfied
		// So we can break out of the loop
		if pq.comp(pq.heap[parent].value, pq.heap[i].value) {
			break
		}

		// If the current node has higher priority (smaller value) than its parent,
		// swap the two nodes to restore the heap property
		pq.swap(i, parent)

		// Move up to the parent's position and continue the process
		i = parent
	}
}

// down restores the heap property by "sifting down" the element at index i.
// It compares the element at i with its children, and if one of the children
// has a higher priority (smaller value for a min-heap), it swaps the elements
// and continues this process recursively down the heap until the heap property is restored.
//
// This function is typically used after removing or modifying an element at the root of the heap,
// to ensure the heap property is maintained throughout the structure.
func (pq *IndexedPriorityQueue[T]) down(i int) {
	n := len(pq.heap) // Get the total number of elements in the heap
	for {
		left := 2*i + 1
		right := 2*i + 2
		smallest := i

		// Check if the left child exists and has a higher priority (smaller value) than the current node
		if left < n && pq.comp(pq.heap[left].value, pq.heap[smallest].value) {
			smallest = left
		}

		// Check if the right child exists and has a higher priority than the smallest value so far
		if right < n && pq.comp(pq.heap[right].value, pq.heap[smallest].value) {
			smallest = right
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
