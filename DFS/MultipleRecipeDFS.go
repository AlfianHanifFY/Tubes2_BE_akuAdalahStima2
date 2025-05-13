package dfs

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

func CountNodes(tree Element.Tree) int64 {
	count := int64(1) // Count the root
	for _, child := range tree.Children {
		count += CountNodes(child)
	}
	return count
}
func cloneMap(original map[string]bool) map[string]bool {
	copy := make(map[string]bool)
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

func buildTreesControlled(root string, recipeMap map[string][]Element.Element, visited map[string]bool, tierLimit int, limit int, nodesVisited *int64) []Element.Tree {
	atomic.AddInt64(nodesVisited, 1)

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
	visited[root] = true
	defer func() { visited[root] = false }()

	for _, recipe := range recipes {
		tierInt := Element.ParseTier(recipe.Tier)
		if tierInt >= tierLimit {
			continue
		}

		left := strings.ToLower(recipe.Left)
		right := strings.ToLower(recipe.Right)

		var leftTrees, rightTrees []Element.Tree
		var wg sync.WaitGroup

		leftChan := make(chan []Element.Tree, 1)
		rightChan := make(chan []Element.Tree, 1)

		wg.Add(2)

		// Proses kiri secara paralel
		go func() {
			defer wg.Done()
			leftChan <- buildTreesControlled(left, recipeMap, cloneMap(visited), tierInt, limit, nodesVisited)
		}()

		// Proses kanan setelah dapat hasil kiri
		go func() {
			defer wg.Done()
			leftResult := <-leftChan
			leftTrees = leftResult
			if len(leftResult) == 0 {
				rightChan <- nil
				return
			}
			rightLimit := int(math.Ceil(float64(limit) / float64(len(leftResult))))
			rightChan <- buildTreesControlled(right, recipeMap, cloneMap(visited), tierInt, rightLimit, nodesVisited)
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

	if len(result) > limit {
		result = result[:limit]
	}
	return result
}

func MultipleRecipeConcurrent(name string, recipeMap map[string][]Element.Element, count int) ([]Element.Tree, MetricsResult) {
	startTime := time.Now()
	var nodesVisited int64 = 0
	var x bool
	name = strings.ToLower(name)
	var trees []Element.Tree
	if Element.IsBaseComponent(name) {
		x = true
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
	} else {
		x = false
		trees = buildTreesControlled(name, recipeMap, map[string]bool{}, math.MaxInt32, count, &nodesVisited)
	}

	if len(trees) > count {
		trees = trees[:count]
	}
	if x {
		duration := time.Since(startTime)
		metrics := MetricsResult{
			NodesVisited:  1,
			Duration:      duration.Milliseconds(),
			DurationHuman: duration.String()}
		return trees, metrics
	} else {
		duration := time.Since(startTime)
		metrics := MetricsResult{
			NodesVisited:  nodesVisited,
			Duration:      duration.Milliseconds(),
			DurationHuman: duration.String()}
		return trees, metrics
	}
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
