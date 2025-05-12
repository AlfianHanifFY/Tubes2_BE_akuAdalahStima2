package bfs

import (
	"fmt"
	"math"
	"stima-2-be/Element"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type MetricsResult struct {
	NodesVisited  int64  `json:"nodes_visited"`
	Duration      int64  `json:"duration_ms"`
	DurationHuman string `json:"duration_human"`
}

type Queue struct {
	items []string
}

func (q *Queue) Enqueue(item string) {
	q.items = append(q.items, item)
}

func (q *Queue) Dequeue() string {
	if len(q.items) == 0 {
		return ""
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q *Queue) IsEmpty() bool {
	return len(q.items) == 0
}

func cloneMap(original map[string]bool) map[string]bool {
	copy := make(map[string]bool)
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

func buildTreesBFS(root string, recipeMap map[string][]Element.Element, visited map[string]bool, tierLimit int, limit int, nodesVisited *int64) []Element.Tree {
	atomic.AddInt64(nodesVisited, 1)

	fmt.Printf("[DEBUG] Visiting: %s (Nodes Visited: %d)\n", root, *nodesVisited)

	if Element.IsBaseComponent(root) {
		return []Element.Tree{
			{
				Root: Element.Element{
					Root:  root,
					Left:  "",
					Right: "",
					Tier:  "0",
				},
				Children: nil,
			},
		}
	}

	if visited[root] {
		return nil
	}

	recipes, exists := recipeMap[strings.ToLower(root)]
	if !exists || root == "time" {
		return nil
	}

	var result []Element.Tree
	queue := &Queue{}
	queue.Enqueue(root)

	visited = cloneMap(visited)
	visited[root] = true

	for !queue.IsEmpty() {
		current := queue.Dequeue()
		recipes, _ = recipeMap[strings.ToLower(current)]

		fmt.Printf("[DEBUG] Current element: %s\n", current)

		// Process all recipes for the current element (root at each level)
		for _, recipe := range recipes {
			tierInt := Element.ParseTier(recipe.Tier)
			if tierInt >= tierLimit {
				fmt.Printf("[DEBUG] Skipping %s due to tier limit\n", recipe.Tier)
				continue
			}

			left := strings.ToLower(recipe.Left)
			right := strings.ToLower(recipe.Right)

			var leftTrees, rightTrees []Element.Tree
			var wg sync.WaitGroup

			leftChan := make(chan []Element.Tree, 1)
			rightChan := make(chan []Element.Tree, 1)

			wg.Add(2)

			// Goroutine untuk kiri
			go func() {
				defer wg.Done()
				leftChan <- buildTreesBFS(left, recipeMap, cloneMap(visited), tierInt, limit, nodesVisited)
			}()

			// Goroutine untuk kanan
			go func() {
				defer wg.Done()
				// Tunggu hasil kiri dulu untuk hitung right limit
				lefts := <-leftChan
				if len(lefts) == 0 {
					rightChan <- nil
					leftTrees = lefts
					return
				}
				rightLimit := int(math.Ceil(float64(limit) / float64(len(lefts))))
				rightChan <- buildTreesBFS(right, recipeMap, cloneMap(visited), tierInt, rightLimit, nodesVisited)
				leftTrees = lefts
			}()

			wg.Wait()
			rightTrees = <-rightChan

			if len(leftTrees) == 0 || len(rightTrees) == 0 {
				continue
			}

			for _, lt := range leftTrees {
				for _, rt := range rightTrees {
					tree := Element.Tree{
						Root:     recipe,
						Children: []Element.Tree{lt, rt},
					}
					result = append(result, tree)
					if len(result) >= limit {
						return result
					}
				}
			}
		}

		// Enqueue the child nodes for BFS traversal
		for _, recipe := range recipes {
			left := strings.ToLower(recipe.Left)
			right := strings.ToLower(recipe.Right)
			if !visited[left] {
				queue.Enqueue(left)
				visited[left] = true
			}
			if !visited[right] {
				queue.Enqueue(right)
				visited[right] = true
			}
		}
	}

	return result
}

func MultipleRecipeConcurrent(name string, recipeMap map[string][]Element.Element, count int) ([]Element.Tree, MetricsResult) {
	startTime := time.Now()
	var nodesVisited int64 = 0

	name = strings.ToLower(name)
	var trees []Element.Tree
	if Element.IsBaseComponent(name) {
		trees = []Element.Tree{
			{
				Root: Element.Element{
					Root:  name,
					Left:  "",
					Right: "",
					Tier:  "0",
				},
				Children: nil,
			},
		}
		nodesVisited = 1
	} else {
		trees = buildTreesBFS(name, recipeMap, map[string]bool{}, math.MaxInt32, count, &nodesVisited)
	}

	if len(trees) > count {
		trees = trees[:count]
	}
	duration := time.Since(startTime)
	metrics := MetricsResult{
		NodesVisited:  nodesVisited,
		Duration:      duration.Milliseconds(),
		DurationHuman: duration.String(),
	}
	return trees, metrics
}

func MultipleRecipe(name string, recipeMap map[string][]Element.Element, count int) ([]Element.Tree, MetricsResult) {
	return MultipleRecipeConcurrent(name, recipeMap, count)
}

func PrintTree(t Element.Tree, indent string) {
	fmt.Printf("%s%s (Tier: %s)\n", indent, t.Root.Root, t.Root.Tier)
	if len(t.Children) > 0 {
		for _, child := range t.Children {
			PrintTree(child, indent+"  ")
		}
	}
}
