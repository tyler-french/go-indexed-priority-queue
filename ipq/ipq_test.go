package ipq_test

import (
	"math/rand"
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
	// Adjacency List
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

// GenerateGraph creates a large random weighted graph with N nodes and E edges.
func GenerateGraph(N, E int) map[int]map[int]int {
	graph := make(map[int]map[int]int, N)

	for i := 0; i < N; i++ {
		graph[i] = make(map[int]int)
	}

	for i := 0; i < E; i++ {
		from := rand.Intn(N)
		to := rand.Intn(N)
		if from != to {
			weight := rand.Intn(100) + 1
			graph[from][to] = weight
		}
	}

	return graph
}

// BenchmarkDijkstra tests the algorithm on large graphs.
func BenchmarkDijkstra(b *testing.B) {
	const N = 10000 // Number of nodes
	const E = 50000 // Number of edges

	graph := GenerateGraph(N, E)
	source := rand.Intn(N)

	b.ResetTimer() // Start timing
	for i := 0; i < b.N; i++ {
		_ = Dijkstra(graph, source) // Run the algorithm
	}
}
