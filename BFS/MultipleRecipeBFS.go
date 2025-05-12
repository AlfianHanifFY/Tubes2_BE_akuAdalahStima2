package bfs

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"stima-2-be/Element"
)

// MetricsResult holds the performance metrics for the tree building process
type MetricsResult struct {
	NodesVisited  int64  `json:"nodes_visited"`
	Duration      int64  `json:"duration_ms"`
	DurationHuman string `json:"duration_human"`
}

func BuildAllValidTreesFIFO(root string, recipeMap map[string][]Element.Element, targetTier int, nodesVisited *int64) []Element.Tree {
	type Node struct {
		Root    string
		Tier    int
		Visited map[string]bool
	}

	runtime.GOMAXPROCS(runtime.NumCPU()) // Use all available cores

	var results []Element.Tree
	var mu sync.Mutex                    // Protects access to `results`
	queue := []Node{
		{
			Root:    root,
			Tier:    targetTier,
			Visited: map[string]bool{},
		},
	}

	var wg sync.WaitGroup
	// Protects access to queue if needed later (not used here)

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		atomic.AddInt64(nodesVisited, 1)

		normalizedRoot := strings.ToLower(curr.Root)
		if Element.IsBaseComponent(normalizedRoot) {
			mu.Lock()
			results = append(results, Element.Tree{
				Root: Element.Element{
					Root:  curr.Root,
					Left:  "",
					Right: "",
					Tier:  "0",
				},
				Children: nil,
			})
			mu.Unlock()
			continue
		}

		if curr.Visited[normalizedRoot] {
			continue
		}

		recipes, exists := recipeMap[normalizedRoot]
		if !exists {
			continue
		}

		for _, recipe := range recipes {
			tierInt := Element.ParseTier(recipe.Tier)
			if tierInt >= curr.Tier {
				continue
			}

			visitedCopy := make(map[string]bool)
			for k, v := range curr.Visited {
				visitedCopy[k] = v
			}
			visitedCopy[normalizedRoot] = true

			wg.Add(1)
			go func(recipe Element.Element, visited map[string]bool) {
				defer wg.Done()

				// Local counter for each thread
				var localVisited int64 = 0

				leftSubTrees := BuildAllValidTreesFIFO(recipe.Left, recipeMap, Element.ParseTier(recipe.Tier), &localVisited)
				rightSubTrees := BuildAllValidTreesFIFO(recipe.Right, recipeMap, Element.ParseTier(recipe.Tier), &localVisited)

				if len(leftSubTrees) == 0 || len(rightSubTrees) == 0 {
					return
				}

				localTrees := make([]Element.Tree, 0, len(leftSubTrees)*len(rightSubTrees))
				for _, lt := range leftSubTrees {
					for _, rt := range rightSubTrees {
						tree := Element.Tree{
							Root: Element.Element{
								Root:  recipe.Root,
								Left:  recipe.Left,
								Right: recipe.Right,
								Tier:  recipe.Tier,
							},
							Children: []Element.Tree{lt, rt},
						}
						localTrees = append(localTrees, tree)
					}
				}

				// Append to shared result with mutex
				mu.Lock()
				results = append(results, localTrees...)
				mu.Unlock()

				atomic.AddInt64(nodesVisited, localVisited)
			}(recipe, visitedCopy)
		}
	}

	wg.Wait()
	return results
}

func MultipleRecipesBFSFIFO(name string, recipeMap map[string][]Element.Element, count int) ([]Element.Tree, MetricsResult) {
	startTime := time.Now()
	var nodesVisited int64 = 0

	normalizedName := strings.ToLower(name)

	if Element.IsBaseComponent(normalizedName) {
		atomic.AddInt64(&nodesVisited, 1)
		duration := time.Since(startTime)
		return []Element.Tree{{
			Root: Element.Element{
				Root:  name,
				Left:  "",
				Right: "",
				Tier:  "0",
			},
			Children: nil,
		}}, MetricsResult{
			NodesVisited:  nodesVisited,
			Duration:      duration.Milliseconds(),
			DurationHuman: duration.String(),
		}
	}

	trees := BuildAllValidTreesFIFO(name, recipeMap, 9999, &nodesVisited)
	if len(trees) > count {
		trees = trees[:count]
	}

	duration := time.Since(startTime)
	return trees, MetricsResult{
		NodesVisited:  nodesVisited,
		Duration:      duration.Milliseconds(),
		DurationHuman: duration.String(),
	}
}

func MultipleRecipe(name string, recipeMap map[string][]Element.Element, count int) ([]Element.Tree, MetricsResult) {
	return MultipleRecipesBFSFIFO(name, recipeMap, count)
}

func PrintTree(t Element.Tree, indent string) {
	fmt.Printf("%s%s (Tier: %s)\n", indent, t.Root.Root, t.Root.Tier)
	if len(t.Children) > 0 {
		for _, child := range t.Children {
			PrintTree(child, indent+"  ")
		}
	}
}
