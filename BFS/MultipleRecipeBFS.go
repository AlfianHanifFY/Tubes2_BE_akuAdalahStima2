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
// Added a new parameter globalVisited to prevent redundant processing
func buildTreeBFS(root string, recipeMap map[string][]Element.Element, branchVisited map[string]bool, globalVisited map[string]bool, nodesVisited *int64) Element.Tree {
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

	// Check if already visited in this branch or globally processed
	// Use normalized name for the visited maps
	if branchVisited[normalizedRoot] || globalVisited[normalizedRoot] {
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

	// Mark as visited in this branch to prevent cycles - use normalized name
	branchVisited[normalizedRoot] = true
	defer func() { branchVisited[normalizedRoot] = false }() // Unmark when done with this branch
	
	// Mark as globally visited to prevent redundant processing
	globalVisited[normalizedRoot] = true

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

	parentTier := getTierAsInt(recipe.Tier)

	// Create node for current element
	current := Element.Element{
		Root:  root, // Keep original casing for display
		Left:  recipe.Left,
		Right: recipe.Right,
		Tier:  recipe.Tier,
	}

	// Process left and right children using BFS approach
	// Only build child if it passes tier validation
	var leftTree, rightTree Element.Tree
	
	// Process left child
	leftNormalized := normalizeElementName(recipe.Left)
	if !Element.IsBaseComponent(leftNormalized) && !branchVisited[leftNormalized] {
		// Check tier first if available
		leftChildTier := -1
		if leftRecipes, exists := recipeMap[leftNormalized]; exists && len(leftRecipes) > 0 {
			leftChildTier = getTierAsInt(leftRecipes[0].Tier)
		}
		
		// Only build if tier validation passes or tier is unknown
		if leftChildTier == -1 || leftChildTier < parentTier {
			leftTree = buildTreeBFS(recipe.Left, recipeMap, branchVisited, globalVisited, nodesVisited)
		} else {
			// If tier validation fails, create a simpler tree node
			leftTree = Element.Tree{
				Root: Element.Element{
					Root:  recipe.Left,
					Left:  "",
					Right: "",
					Tier:  strconv.Itoa(leftChildTier),
				},
				Children: nil,
			}
			fmt.Printf("TIER SKIPPED: %s (Tier %d) has child %s (Tier %d) with equal/higher tier\n", 
				root, parentTier, recipe.Left, leftChildTier)
		}
	} else {
		// For base components or already visited branches
		leftTree = Element.Tree{
			Root: Element.Element{
				Root:  recipe.Left,
				Left:  "",
				Right: "",
				Tier:  func() string {
					if Element.IsBaseComponent(leftNormalized) {
						return "0"
					}
					return ""
				}(),
			},
			Children: nil,
		}
	}
	
	// Process right child
	rightNormalized := normalizeElementName(recipe.Right)
	if !Element.IsBaseComponent(rightNormalized) && !branchVisited[rightNormalized] {
		// Check tier first if available
		rightChildTier := -1
		if rightRecipes, exists := recipeMap[rightNormalized]; exists && len(rightRecipes) > 0 {
			rightChildTier = getTierAsInt(rightRecipes[0].Tier)
		}
		
		// Only build if tier validation passes or tier is unknown
		if rightChildTier == -1 || rightChildTier < parentTier {
			rightTree = buildTreeBFS(recipe.Right, recipeMap, branchVisited, globalVisited, nodesVisited)
		} else {
			// If tier validation fails, create a simpler tree node
			rightTree = Element.Tree{
				Root: Element.Element{
					Root:  recipe.Right,
					Left:  "",
					Right: "",
					Tier:  strconv.Itoa(rightChildTier),
				},
				Children: nil,
			}
			fmt.Printf("TIER SKIPPED: %s (Tier %d) has child %s (Tier %d) with equal/higher tier\n", 
				root, parentTier, recipe.Right, rightChildTier)
		}
	} else {
		// For base components or already visited branches
		rightTree = Element.Tree{
			Root: Element.Element{
				Root:  recipe.Right,
				Left:  "",
				Right: "",
				Tier:  func() string {
					if Element.IsBaseComponent(rightNormalized) {
						return "0"
					}
					return ""
				}(),
			},
			Children: nil,
		}
	}

	// Create tree with this recipe as root
	result := Element.Tree{
		Root:     current,
		Children: []Element.Tree{leftTree, rightTree},
	}

	return result
}

// MultipleRecipesBFS returns multiple recipe trees using BFS with multithreading support
// Now with tier validation, redundancy prevention and metrics tracking
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
	globalProcessedMap := make(map[string]bool) // Shared map to prevent redundant processing

	// Process each recipe in parallel
	for i := 0; i < now; i++ {
		wg.Add(1)
		
		// Launch a goroutine for each recipe
		go func(recipe Element.Element, index int) {
			defer wg.Done()
			
			// Create new visited map for this branch to track cycles
			branchVisited := make(map[string]bool) 
			
			// Create thread-local copy of global processed map
			var localGlobalProcessed map[string]bool
			mutex.Lock()
			localGlobalProcessed = make(map[string]bool, len(globalProcessedMap))
			for k, v := range globalProcessedMap {
				localGlobalProcessed[k] = v
			}
			mutex.Unlock()

			rootElement := Element.Element{
				Root:  name,
				Left:  recipe.Left,
				Right: recipe.Right,
				Tier:  recipe.Tier,
			}

			// Increment nodes visited counter for root
			atomic.AddInt64(&nodesVisited, 1)
			
			parentTier := getTierAsInt(recipe.Tier)

			// Process left child with tier validation
			leftNormalized := normalizeElementName(recipe.Left)
			var leftTree Element.Tree
			if !Element.IsBaseComponent(leftNormalized) {
				// Check tier first if available
				leftChildTier := -1
				if leftRecipes, exists := normalizedRecipeMap[leftNormalized]; exists && len(leftRecipes) > 0 {
					leftChildTier = getTierAsInt(leftRecipes[0].Tier)
				}
				
				// Only build if tier validation passes or tier is unknown
				if leftChildTier == -1 || leftChildTier < parentTier {
					leftTree = buildTreeBFS(recipe.Left, normalizedRecipeMap, branchVisited, localGlobalProcessed, &nodesVisited)
				} else {
					// If tier validation fails, create a simpler tree node
					leftTree = Element.Tree{
						Root: Element.Element{
							Root:  recipe.Left,
							Left:  "",
							Right: "",
							Tier:  strconv.Itoa(leftChildTier),
						},
						Children: nil,
					}
					fmt.Printf("TIER SKIPPED: %s (Tier %d) has child %s (Tier %d) with equal/higher tier\n", 
						name, parentTier, recipe.Left, leftChildTier)
				}
			} else {
				// For base components
				leftTree = Element.Tree{
					Root: Element.Element{
						Root:  recipe.Left,
						Left:  "",
						Right: "",
						Tier:  "0",
					},
					Children: nil,
				}
			}
			
			// Reset branch visited for right side
			branchVisited = make(map[string]bool)
			
			// Process right child with tier validation
			rightNormalized := normalizeElementName(recipe.Right)
			var rightTree Element.Tree
			if !Element.IsBaseComponent(rightNormalized) {
				// Check tier first if available
				rightChildTier := -1
				if rightRecipes, exists := normalizedRecipeMap[rightNormalized]; exists && len(rightRecipes) > 0 {
					rightChildTier = getTierAsInt(rightRecipes[0].Tier)
				}
				
				// Only build if tier validation passes or tier is unknown
				if rightChildTier == -1 || rightChildTier < parentTier {
					rightTree = buildTreeBFS(recipe.Right, normalizedRecipeMap, branchVisited, localGlobalProcessed, &nodesVisited)
				} else {
					// If tier validation fails, create a simpler tree node
					rightTree = Element.Tree{
						Root: Element.Element{
							Root:  recipe.Right,
							Left:  "",
							Right: "",
							Tier:  strconv.Itoa(rightChildTier),
						},
						Children: nil,
					}
					fmt.Printf("TIER SKIPPED: %s (Tier %d) has child %s (Tier %d) with equal/higher tier\n", 
						name, parentTier, recipe.Right, rightChildTier)
				}
			} else {
				// For base components
				rightTree = Element.Tree{
					Root: Element.Element{
						Root:  recipe.Right,
						Left:  "",
						Right: "",
						Tier:  "0",
					},
					Children: nil,
				}
			}

			// Create tree with this recipe as root
			tree := Element.Tree{
				Root:     rootElement,
				Children: []Element.Tree{leftTree, rightTree},
			}
			
			// Update global processed map with the new processed elements
			mutex.Lock()
			for k, v := range localGlobalProcessed {
				if v {
					globalProcessedMap[k] = true
				}
			}
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