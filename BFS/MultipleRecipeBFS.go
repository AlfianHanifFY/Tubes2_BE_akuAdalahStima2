package bfs

import (
	"fmt"
	"strings"
	"stima-2-be/Element"
	"sync"
	"strconv"
)

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
// Added tier validation
func buildTreeBFS(root string, recipeMap map[string][]Element.Element, visited map[string]bool) Element.Tree {
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
	leftTree := buildTreeBFS(recipe.Left, recipeMap, visited)
	rightTree := buildTreeBFS(recipe.Right, recipeMap, visited)

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
// Now with tier validation
func MultipleRecipesBFS(name string, recipeMap map[string][]Element.Element, count int) []Element.Tree {
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
	
	fmt.Printf("Debug: name=%s, normalized=%s, count=%d\n", name, normalizeElementName(name), count)

	if len(recipes) == 0 {
		// If target is a base component, return it as tier 0
		if Element.IsBaseComponent(name) || Element.IsBaseComponent(normalizeElementName(name)) {
			return []Element.Tree{{
				Root: Element.Element{
					Root:  name,
					Left:  "",
					Right: "",
					Tier:  "0",
				},
				Children: nil,
			}}
		}

		// Otherwise return as unknown tier
		return []Element.Tree{{
			Root: Element.Element{
				Root:  name,
				Left:  "",
				Right: "",
				Tier:  "",
			},
			Children: nil,
		}}
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

			// Build left and right subtrees using BFS
			leftTree := buildTreeBFS(recipe.Left, normalizedRecipeMap, visited)
			rightTree := buildTreeBFS(recipe.Right, normalizedRecipeMap, visited)

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

	// Validate trees to ensure all leaf nodes are base components
	var validTrees []Element.Tree
	for _, tree := range results {
		if Element.ValidateTree(tree) {
			validTrees = append(validTrees, tree)
		} else {
			fmt.Printf("Tree is invalid: %v\n", tree.Root)
		}
	}

	if len(validTrees) == 0 {
		fmt.Println("Warning: No valid trees found where all leaf nodes are base components")
		return results // Return all results even if none are valid (same as DFS implementation)
	} else {
		fmt.Printf("Found %d valid trees where all leaf nodes are base components (BFS)\n", len(validTrees))
	}

	return results // Return all results (same as DFS implementation)
}

// SimpleMultipleRecipesBFS is a non-multithreaded version that follows the same pattern as DFS
// Now with tier validation
func SimpleMultipleRecipesBFS(name string, recipeMap map[string][]Element.Element, count int) []Element.Tree {
	var results []Element.Tree
	recipes, exists := recipeMap[name]
	
	fmt.Printf("Debug: name=%s, count=%d\n", name, count)

	if !exists || len(recipes) == 0 {
		// If target is a base component, return it as tier 0
		if Element.IsBaseComponent(name) {
			return []Element.Tree{{
				Root: Element.Element{
					Root:  name,
					Left:  "",
					Right: "",
					Tier:  "0",
				},
				Children: nil,
			}}
		}

		// Otherwise return as unknown tier
		return []Element.Tree{{
			Root: Element.Element{
				Root:  name,
				Left:  "",
				Right: "",
				Tier:  "",
			},
			Children: nil,
		}}
	}

	// Limit the number of recipes we consider
	now := len(recipes)
	if now > count {
		now = count
	}

	for i := 0; i < now; i++ {
		recipe := recipes[i]
		visited := make(map[string]bool)

		rootElement := Element.Element{
			Root:  name,
			Left:  recipe.Left,
			Right: recipe.Right,
			Tier:  recipe.Tier,
		}

		// Build left and right subtrees using BFS
		leftTree := buildTreeBFS(recipe.Left, recipeMap, visited)
		rightTree := buildTreeBFS(recipe.Right, recipeMap, visited)

		// Create tree with this recipe as root
		tree := Element.Tree{
			Root:     rootElement,
			Children: []Element.Tree{leftTree, rightTree},
		}
		
		// Validate tier consistency
		validateTierConsistency(tree)
		
		results = append(results, tree)
	}

	// Validate trees to ensure all leaf nodes are base components
	validTrees := make([]Element.Tree, 0)
	for _, tree := range results {
		if Element.ValidateTree(tree) {
			validTrees = append(validTrees, tree)
		} else {
			fmt.Printf("Tree is invalid: %v\n", tree.Root)
		}
	}

	if len(validTrees) == 0 {
		fmt.Println("Warning: No valid trees found where all leaf nodes are base components")
	} else {
		fmt.Printf("Found %d valid trees where all leaf nodes are base components (BFS)\n", len(validTrees))
	}

	return results // Return all results (same as DFS implementation)
}