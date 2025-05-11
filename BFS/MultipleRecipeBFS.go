package bfs

import (
	"fmt"
	"strings"
	"stima-2-be/Element"
	"sync"
	"sync/atomic"
	"strconv"
	"time"
)

// MetricsResult holds the performance metrics for the tree building process
// This matches the DFS implementation for consistency
type MetricsResult struct {
	NodesVisited  int64  `json:"nodes_visited"`
	Duration      int64  `json:"duration_ms"`
	DurationHuman string `json:"duration_human"`
}

// normalizeElementName normalizes element names to handle case sensitivity
// This ensures that elements like "Water" and "water" are treated the same
func normalizeElementName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// Helper function to convert tier string to integer
func getTierAsInt(tier string) int {
	if tier == "" {
		return -1 // Tier tidak diketahui
	}
	result, err := strconv.Atoi(tier)
	if err != nil {
		return -1 // Tier tidak valid
	}
	return result
}

// Build a tree using BFS approach with case-insensitive element names
// Added tier validation and nodes visited counter
func buildTreeBFS(root string, recipeMap map[string][]Element.Element, visited map[string]bool, nodesVisited *int64) Element.Tree {
	// Increment nodes visited counter
	atomic.AddInt64(nodesVisited, 1)
	
	// Normalize the root name for case-insensitive comparison
	normalizedRoot := normalizeElementName(root)
	
	// If this is a base component, return it as a leaf
	if Element.IsBaseComponent(normalizedRoot) || Element.IsBaseComponent(root) {
		return Element.Tree{
			Root: Element.Element{
				Root:  root, // Keep original casing for display
				Left:  "",
				Right: "",
				Tier:  "0", // Base components = tier 0
			},
			Children: nil,
		}
	}

	// Check if already visited or no recipes available
	// Use normalized name for the visited map
	if visited[normalizedRoot] {
		return Element.Tree{
			Root: Element.Element{
				Root:  root, // Keep original casing for display
				Left:  "",
				Right: "",
				Tier:  "", // Unknown tier
			},
			Children: nil,
		}
	}

	// Mark as visited to prevent cycles - use normalized name
	visited[normalizedRoot] = true
	defer func() { visited[normalizedRoot] = false }() // Unmark when done with this branch

	// Find recipes for the current element - try both original and normalized names
	var recipes []Element.Element
	if r, exists := recipeMap[root]; exists && len(r) > 0 {
		recipes = r
	} else if r, exists := recipeMap[normalizedRoot]; exists && len(r) > 0 {
		recipes = r
	}

	if len(recipes) == 0 {
		return Element.Tree{
			Root: Element.Element{
				Root:  root, // Keep original casing for display
				Left:  "",
				Right: "",
				Tier:  "", // Unknown tier
			},
			Children: nil,
		}
	}

	// Use the first recipe
	recipe := recipes[0]

	// Create node for current element
	current := Element.Element{
		Root:  root, // Keep original casing for display
		Left:  recipe.Left,
		Right: recipe.Right,
		Tier:  recipe.Tier,
	}

	// Process left and right children using BFS approach
	leftTree := buildTreeBFS(recipe.Left, recipeMap, visited, nodesVisited)
	rightTree := buildTreeBFS(recipe.Right, recipeMap, visited, nodesVisited)

	// Create tree with this recipe as root
	result := Element.Tree{
		Root:     current,
		Children: []Element.Tree{leftTree, rightTree},
	}

	// Validate tier consistency
	validateTierConsistency(result)

	return result
}

// Function to validate tier consistency
func validateTierConsistency(tree Element.Tree) bool {
	if (tree.Children == nil) {
		return true
	}

	parentTierStr := tree.Root.Tier
	parentTier := getTierAsInt(parentTierStr)
	consistent := true

	for _, child := range tree.Children {
		// Skip base components
		if child.Root.Tier == "0" {
			continue
		}

		childTier := getTierAsInt(child.Root.Tier)
		
		// Skip if child tier is unknown (-1)
		if childTier == -1 {
			continue
		}

		// Detect tier anomaly: child tier > parent tier
		if childTier > parentTier {
			fmt.Printf("ANOMALI TIER: %s (Tier %s) memiliki child %s (Tier %s) dengan tier lebih tinggi\n", 
				tree.Root.Root, tree.Root.Tier, 
				child.Root.Root, child.Root.Tier)
			consistent = false
		}

		// Recurse for children
		childConsistent := validateTierConsistency(child)
		consistent = consistent && childConsistent
	}

	return consistent
}

// MultipleRecipesBFS returns multiple recipe trees using BFS with multithreading support
// Now with tier validation and metrics tracking like the DFS implementation
func MultipleRecipesBFS(name string, recipeMap map[string][]Element.Element, count int) ([]Element.Tree, MetricsResult) {
	startTime := time.Now()
	var nodesVisited int64 = 0
	
	// Create case-insensitive recipe map
	normalizedRecipeMap := make(map[string][]Element.Element)
	
	// Populate the normalized recipe map
	for key, recipes := range recipeMap {
		normalizedKey := normalizeElementName(key)
		normalizedRecipeMap[normalizedKey] = recipes
		// Keep the original key as well for direct lookups
		normalizedRecipeMap[key] = recipes
	}
	
	var results []Element.Tree
	
	// Try to find recipes using both original and normalized names
	var recipes []Element.Element
	if r, exists := recipeMap[name]; exists && len(r) > 0 {
		recipes = r
	} else {
		normalizedName := normalizeElementName(name)
		if r, exists := recipeMap[normalizedName]; exists && len(r) > 0 {
			recipes = r
		}
	}

	if len(recipes) == 0 {
		// If target is a base component, return it as tier 0
		if Element.IsBaseComponent(name) || Element.IsBaseComponent(normalizeElementName(name)) {
			atomic.AddInt64(&nodesVisited, 1)
			duration := time.Since(startTime)
			metrics := MetricsResult{
				NodesVisited:  nodesVisited,
				Duration:      duration.Milliseconds(),
				DurationHuman: duration.String(),
			}
			return []Element.Tree{{
				Root: Element.Element{
					Root:  name,
					Left:  "",
					Right: "",
					Tier:  "0",
				},
				Children: nil,
			}}, metrics
		}

		// Otherwise return as unknown tier
		duration := time.Since(startTime)
		metrics := MetricsResult{
			NodesVisited:  nodesVisited,
			Duration:      duration.Milliseconds(),
			DurationHuman: duration.String(),
		}
		return []Element.Tree{{
			Root: Element.Element{
				Root:  name,
				Left:  "",
				Right: "",
				Tier:  "",
			},
			Children: nil,
		}}, metrics
	}

	// Limit the number of recipes we consider
	now := len(recipes)
	if now > count {
		now = count
	}

	// Use WaitGroup for synchronization
	var wg sync.WaitGroup
	var mutex sync.Mutex // To protect concurrent writes to results slice

	// Process each recipe in parallel
	for i := 0; i < now; i++ {
		wg.Add(1)
		
		// Launch a goroutine for each recipe
		go func(recipe Element.Element, index int) {
			defer wg.Done()
			
			visited := make(map[string]bool) // Create new visited map for each recipe

			rootElement := Element.Element{
				Root:  name,
				Left:  recipe.Left,
				Right: recipe.Right,
				Tier:  recipe.Tier,
			}

			// Increment nodes visited counter for root
			atomic.AddInt64(&nodesVisited, 1)

			// Build left and right subtrees using BFS
			leftTree := buildTreeBFS(recipe.Left, normalizedRecipeMap, visited, &nodesVisited)
			rightTree := buildTreeBFS(recipe.Right, normalizedRecipeMap, visited, &nodesVisited)

			// Create tree with this recipe as root
			tree := Element.Tree{
				Root:     rootElement,
				Children: []Element.Tree{leftTree, rightTree},
			}
			
			// Validate tier consistency for this tree
			validateTierConsistency(tree)
			
			// Safely append to results
			mutex.Lock()
			results = append(results, tree)
			mutex.Unlock()
			
		}(recipes[i], i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Calculate metrics
	duration := time.Since(startTime)
	metrics := MetricsResult{
		NodesVisited:  nodesVisited,
		Duration:      duration.Milliseconds(),
		DurationHuman: duration.String(),
	}

	return results, metrics
}

// Wrapper function for backward compatibility, similar to DFS implementation
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
	return MultipleRecipesBFS(name, recipeMap, count)
}

// Simple function to create a debug view of the tree, similar to DFS implementation
func PrintTree(t Element.Tree, indent string) {
	fmt.Printf("%s%s (Tier: %s)\n", indent, t.Root.Root, t.Root.Tier)
	if len(t.Children) > 0 {
		for _, child := range t.Children {
			PrintTree(child, indent+"  ")
		}
	}
}