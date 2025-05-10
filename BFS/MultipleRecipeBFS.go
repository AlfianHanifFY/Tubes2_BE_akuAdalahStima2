package bfs

import (
	"fmt"
	"stima-2-be/Element"
)

// Build a tree using BFS approach
func buildTreeBFS(root string, recipeMap map[string][]Element.Element, visited map[string]bool) Element.Tree {
	// If this is a base component, return it as a leaf
	if Element.IsBaseComponent(root) {
		return Element.Tree{
			Root: Element.Element{
				Root:  root,
				Left:  "",
				Right: "",
				Tier:  "0", // Base components = tier 0
			},
			Children: nil,
		}
	}

	// Check if already visited or no recipes available
	if visited[root] || len(recipeMap[root]) == 0 {
		return Element.Tree{
			Root: Element.Element{
				Root:  root,
				Left:  "",
				Right: "",
				Tier:  "", // Unknown tier
			},
			Children: nil,
		}
	}

	// Mark as visited to prevent cycles
	visited[root] = true
	defer func() { visited[root] = false }() // Unmark when done with this branch

	// Get recipes for the current element
	recipes := recipeMap[root]
	if len(recipes) == 0 {
		return Element.Tree{
			Root: Element.Element{
				Root:  root,
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
	current := Element.Tree{
		Root: Element.Element{
			Root:  root,
			Left:  recipe.Left,
			Right: recipe.Right,
			Tier:  recipe.Tier,
		},
	}

	// Process left and right children using BFS approach
	// We'll track and expand nodes level by level
	leftTree := buildTreeBFS(recipe.Left, recipeMap, visited)
	rightTree := buildTreeBFS(recipe.Right, recipeMap, visited)

	// Set children
	current.Children = []Element.Tree{leftTree, rightTree}

	return current
}

func MultipleRecipesBFS(name string, count int) []Element.Tree {
	recipeMap := Element.BuildRecipeMap()
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
		visited := make(map[string]bool) // Create new visited map for each recipe

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
		return results // Return all results even if none are valid (same as DFS implementation)
	} else {
		fmt.Printf("Found %d valid trees where all leaf nodes are base components (BFS)\n", len(validTrees))
	}

	return results // Return all results (same as DFS implementation)
}