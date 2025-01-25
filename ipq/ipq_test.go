package ipq_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tyler-french/go-indexed-priority-queue/ipq"
)

// Dijkstra finds shortest paths from `src` in a weighted graph.
func Dijkstra(graph map[int]map[int]int, src int) map[int]int {
	pq := ipq.NewInt()

	dist := make(map[int]int)
	for node := range graph {
		dist[node] = 1<<31 - 1
	}

	dist[src] = 0
	pq.Push(src, 0)

	for !pq.Empty() {
		node, d, _ := pq.Pop()
		for neighbor, weight := range graph[node] {
			newDist := d + weight
			if newDist < dist[neighbor] { // Found a shorter path
				dist[neighbor] = newDist
				pq.Push(neighbor, newDist)
			}
		}
	}
	return dist
}

func TestDijkstra(t *testing.T) {
	graph := map[int]map[int]int{
		0: {1: 4, 2: 1},
		1: {3: 1},
		2: {1: 2, 3: 5},
		3: {},
	}

	result := Dijkstra(graph, 0)

	// Expected shortest distances
	expected := map[int]int{
		0: 0, // Source
		1: 3, // 0 → 2 → 1 (1 + 2)
		2: 1, // 0 → 2 (1)
		3: 4, // 0 → 2 → 1 → 3 (1 + 2 + 1)
	}

	assert.Equal(t, expected, result, "Dijkstra’s output is incorrect")
}
