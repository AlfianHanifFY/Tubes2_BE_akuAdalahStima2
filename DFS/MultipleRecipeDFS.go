package dfs

import (
	"fmt"
	"runtime"
	"stima-2-be/Element"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// MetricsResult holds the performance metrics for the tree building process
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

// Build a tree with depth limitation and ensuring base components as leaves
// This maintains the original logic but is used for recursive calls
func buildAllValidTrees(root string, recipeMap map[string][]Element.Element, visited map[string]bool, targetTier int, nodesVisited *int64) []Element.Tree {
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

	var trees []Element.Tree
	visited[root] = true
	defer func() { visited[root] = false }()

	for _, recipe := range recipes {
		// Parse tier and filter if tier >= targetTier
		tierInt := Element.ParseTier(recipe.Tier)
		if tierInt >= targetTier {
			continue
		}

		left := strings.ToLower(recipe.Left)
		right := strings.ToLower(recipe.Right)

		leftSubTrees := buildAllValidTrees(left, recipeMap, visited, tierInt, nodesVisited)
		rightSubTrees := buildAllValidTrees(right, recipeMap, visited, tierInt, nodesVisited)
		if leftSubTrees != nil && rightSubTrees != nil && len(leftSubTrees) > 0 && len(rightSubTrees) > 0 {
			// This node is successful in building valid trees
			for _, left := range leftSubTrees {
				for _, right := range rightSubTrees {
					// atomic.AddInt64(nodesVisited, 1)
					tree := Element.Tree{
						Root: Element.Element{
							Root:  root,
							Left:  recipe.Left,
							Right: recipe.Right,
							Tier:  recipe.Tier,
						},
						Children: []Element.Tree{left, right},
					}
					trees = append(trees, tree)
				}
			}
		}
	}
	return trees
}

// The multithreaded version of MultipleRecipe that preserves tier behavior
func MultipleRecipeConcurrent(name string, recipeMap map[string][]Element.Element, count int) ([]Element.Tree, MetricsResult) {
	startTime := time.Now()
	var nodesVisited int64 = 0

	name = strings.ToLower(name)
	recipes, exists := recipeMap[name]
	if !exists {
		if Element.IsBaseComponent(name) {
			duration := time.Since(startTime)
			metrics := MetricsResult{
				NodesVisited:  nodesVisited,
				Duration:      duration.Milliseconds(),
				DurationHuman: duration.String(),
			}
			return []Element.Tree{
				{
					Root: Element.Element{
						Root:  name,
						Left:  "",
						Right: "",
						Tier:  "0",
					},
					Children: nil,
				},
			}, metrics
		}
		duration := time.Since(startTime)
		metrics := MetricsResult{
			NodesVisited:  nodesVisited,
			Duration:      duration.Milliseconds(),
			DurationHuman: duration.String(),
		}
		return nil, metrics
	}

	var allTrees []Element.Tree
	var allTreesMutex sync.Mutex

	maxWorkers := runtime.NumCPU() / 2
	if maxWorkers < 1 {
		maxWorkers = 1
	}

	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	var done int32 = 0

	for _, recipe := range recipes {
		if atomic.LoadInt32(&done) == 1 {
			break
		}

		tier := Element.ParseTier(recipe.Tier)
		left := strings.ToLower(recipe.Left)
		right := strings.ToLower(recipe.Right)

		// Buat visited map baru untuk setiap panggilan
		leftSubTrees := buildAllValidTrees(left, recipeMap, map[string]bool{}, tier, &nodesVisited)
		rightSubTrees := buildAllValidTrees(right, recipeMap, map[string]bool{}, tier, &nodesVisited)

		if leftSubTrees == nil || rightSubTrees == nil || len(leftSubTrees) == 0 || len(rightSubTrees) == 0 {
			continue
		}

		for _, leftTree := range leftSubTrees {
			for _, rightTree := range rightSubTrees {
				if atomic.LoadInt32(&done) == 1 {
					break
				}

				semaphore <- struct{}{}
				wg.Add(1)

				go func(lt, rt Element.Tree, rec Element.Element) {
					defer func() {
						<-semaphore
						wg.Done()
					}()

					tree := Element.Tree{
						Root: Element.Element{
							Root:  rec.Root,
							Left:  rec.Left,
							Right: rec.Right,
							Tier:  rec.Tier,
						},
						Children: []Element.Tree{lt, rt},
					}

					allTreesMutex.Lock()
					defer allTreesMutex.Unlock()

					if int(atomic.LoadInt32(&done)) == 0 {
						allTrees = append(allTrees, tree)
						// Hanya tambah node jika tree valid ditambahkan
						nodeCount := CountNodes(tree)
						atomic.AddInt64(&nodesVisited, nodeCount)

						if len(allTrees) >= count {
							atomic.StoreInt32(&done, 1)
						}
					}
				}(leftTree, rightTree, recipe)
			}
		}
	}

	wg.Wait()

	// Potong jika jumlah melebihi batas count karena race di akhir
	if len(allTrees) > count {
		allTrees = allTrees[:count]
	}

	duration := time.Since(startTime)

	metrics := MetricsResult{
		NodesVisited:  nodesVisited,
		Duration:      duration.Milliseconds(),
		DurationHuman: duration.String(),
	}

	return allTrees, metrics
}

// Wrapper function for backward compatibility
func MultipleRecipe(name string, recipeMap map[string][]Element.Element, count int) ([]Element.Tree, MetricsResult) {
	if Element.IsBaseComponent(name) {
		startTime := time.Now()
		duration := time.Since(startTime)
		metrics := MetricsResult{
			NodesVisited:  1,
			Duration:      duration.Milliseconds(),
			DurationHuman: duration.String(),
		}
		return []Element.Tree{
			{
				Root: Element.Element{
					Root:  name,
					Left:  "",
					Right: "",
					Tier:  "0",
				},
				Children: nil,
			},
		}, metrics
	}
	return MultipleRecipeConcurrent(name, recipeMap, count)
}

// Simple function to create a debug view of the tree
func PrintTree(t Element.Tree, indent string) {
	fmt.Printf("%s%s (Tier: %s)\n", indent, t.Root.Root, t.Root.Tier)
	if len(t.Children) > 0 {
		for _, child := range t.Children {
			PrintTree(child, indent+"  ")
		}
	}
}

// func main() {
// 	// Parse command line arguments
// 	if len(os.Args) < 2 {
// 		fmt.Println("Usage: program <filename> [target] [maxResults]")
// 		fmt.Println("Example: program akuAdalahStima.json Dust 3")
// 		os.Exit(1)
// 	}

// 	filename := os.Args[1]
// 	target := "Dust" // Default target
// 	maxResults := 3  // Default max results

// 	if len(os.Args) > 2 {
// 		target = os.Args[2]
// 	}

// 	if len(os.Args) > 3 {
// 		fmt.Sscanf(os.Args[3], "%d", &maxResults)
// 	}

// 	// Add debug mode flag
// 	debugMode := false
// 	if len(os.Args) > 4 && strings.ToLower(os.Args[4]) == "debug" {
// 		debugMode = true
// 	}

// 	fmt.Println("Loading recipes from:", filename)
// 	fmt.Println("Target:", target)
// 	fmt.Println("Max results:", maxResults)

// 	err := Element.LoadElementsFromFile(filename)
// 	if err != nil {
// 		fmt.Println("Error loading recipes:", err)
// 		os.Exit(1)
// 	}

// 	fmt.Println("Loaded", len(Element.GetAllElement()), "recipes")
// 	recipeMap := Element.BuildRecipeMap(Element.GetAllElement())

// 	// Use the optimized approach to find combinations
// 	fmt.Println("Building recipe trees...")
// 	trees, metrics := MultipleRecipe(target, recipeMap, maxResults)

// 	// Print metrics
// 	fmt.Printf("Nodes visited: %d\n", metrics.NodesVisited)
// 	fmt.Printf("Duration: %s (%d ms)\n", metrics.DurationHuman, metrics.Duration)

// 	// Validate trees
// 	// validTrees := make([]Element.Tree, 0)
// 	// for _, tree := range trees {
// 	// 	if Element.ValidateTree(tree) {
// 	// 		validTrees = append(validTrees, tree)
// 	// 	}
// 	// }

// 	// if len(validTrees) == 0 {
// 	// 	fmt.Println("Warning: No valid trees found where all leaf nodes are base components")
// 	// 	// Fall back to original trees
// 	// 	validTrees = trees
// 	// } else {
// 	// 	fmt.Printf("Found %d valid trees where all leaf nodes are base components\n", len(validTrees))
// 	// }

// 	if debugMode {
// 		// Print tree structure for debugging
// 		fmt.Println("\nTree structure:")
// 		for i, tree := range trees {
// 			fmt.Printf("Tree %d:\n", i+1)
// 			PrintTree(tree, "")
// 			fmt.Println()
// 		}
// 	} else {
// 		// Create a response object with both trees and metrics
// 		type Response struct {
// 			Trees   []Element.Tree `json:"trees"`
// 			Metrics MetricsResult  `json:"metrics"`
// 		}

// 		response := Response{
// 			Trees:   trees,
// 			Metrics: metrics,
// 		}

// 		// Output JSON
// 		output, err := json.MarshalIndent(response, "", "  ")
// 		if err != nil {
// 			fmt.Println("Error encoding result:", err)
// 			return
// 		}

// 		fmt.Println(string(output))
// 	}
// }
