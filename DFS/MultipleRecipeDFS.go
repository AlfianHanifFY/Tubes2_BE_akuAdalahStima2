package dfs

import (
	"fmt"
	"math"
	"stima-2-be/Element"
	"strings"
	"sync/atomic"
	"time"
)

type MetricsResult struct {
	NodesVisited  int64  `json:"nodes_visited"`
	Duration      int64  `json:"duration_ms"`
	DurationHuman string `json:"duration_human"`
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
	if !exists {
		return nil
	}

	if root == "time" {
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

		leftTrees := buildTreesControlled(left, recipeMap, visited, tierInt, limit, nodesVisited)
		if len(leftTrees) == 0 {
			continue
		}

		rightLimit := int(math.Ceil(float64(limit) / float64(len(leftTrees))))
		rightTrees := buildTreesControlled(right, recipeMap, visited, tierInt, rightLimit, nodesVisited)
		if len(rightTrees) == 0 {
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
	} else {
		trees = buildTreesControlled(name, recipeMap, map[string]bool{}, math.MaxInt32, count, &nodesVisited)
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
