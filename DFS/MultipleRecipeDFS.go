package dfs

import (
	"fmt"
	"stima-2-be/Element"
)

// Build a tree with depth limitation and ensuring base components as leaves
func buildTreeWithDepth(root string, recipeMap map[string][]Element.Element, visited map[string]bool, depth int) Element.Tree {
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

	// Base cases: max depth reached, already visited, or no recipes available
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

	// get first recipe
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

	recipe := recipes[0]

	// Build left and right subtrees with increased depth
	leftTree := buildTreeWithDepth(recipe.Left, recipeMap, visited, depth+1)
	rightTree := buildTreeWithDepth(recipe.Right, recipeMap, visited, depth+1)

	// Create the tree with this node as root
	return Element.Tree{
		Root: Element.Element{
			Root:  root,
			Left:  recipe.Left,
			Right: recipe.Right,
			Tier:  recipe.Tier,
		},
		Children: []Element.Tree{leftTree, rightTree},
	}
}

func MultipleRecipe(name string, recipeMap map[string][]Element.Element, count int) []Element.Tree {
	var results []Element.Tree
	recipes, exists := recipeMap[name]

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

		// Build children trees with depth limitation
		leftTree := buildTreeWithDepth(recipe.Left, recipeMap, visited, 1)
		rightTree := buildTreeWithDepth(recipe.Right, recipeMap, visited, 1)

		tree := Element.Tree{
			Root:     rootElement,
			Children: []Element.Tree{leftTree, rightTree},
		}
		results = append(results, tree)
	}

	
	validTrees := make([]Element.Tree, 0)
	for _, tree := range results {
		if Element.ValidateTree(tree) {
			validTrees = append(validTrees, tree)
		}
	}

	if len(validTrees) == 0 {
		fmt.Println("Warning: No valid trees found where all leaf nodes are base components")
	} else {
		fmt.Printf("Found %d valid trees where all leaf nodes are base components\n", len(validTrees))
	}

	return results
}

// // Simple function to create a debug view of the tree
// func printTree(t Element.Tree, indent string) {
// 	fmt.Printf("%s%s (Tier: %s)\n", indent, t.Root.Root, t.Root.Tier)
// 	if len(t.Children) > 0 {
// 		for _, child := range t.Children {
// 			printTree(child, indent+"  ")
// 		}
// 	}
// }

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

// 	recipeMap := Element.BuildRecipeMap(Element.GetAllElement())

// 	// Use the optimized approach to find combinations
// 	fmt.Println("Building recipe trees...")
// 	trees := MultipleRecipe(target, recipeMap, maxResults)

// 	// Validate trees
// 	validTrees := make([]Element.Tree, 0)
// 	for _, tree := range trees {
// 		if Element.ValidateTree(tree) {
// 			validTrees = append(validTrees, tree)
// 		}
// 	}

// 	if len(validTrees) == 0 {
// 		fmt.Println("Warning: No valid trees found where all leaf nodes are base components")
// 		// Fall back to original trees
// 		validTrees = trees
// 	} else {
// 		fmt.Printf("Found %d valid trees where all leaf nodes are base components\n", len(validTrees))
// 	}

// 	if debugMode {
// 		// Print tree structure for debugging
// 		fmt.Println("\nTree structure:")
// 		for i, tree := range validTrees {
// 			fmt.Printf("Tree %d:\n", i+1)
// 			printTree(tree, "")
// 			fmt.Println()
// 		}
// 	} else {
// 		// Output JSON
// 		output, err := json.MarshalIndent(validTrees, "", "  ")
// 		if err != nil {
// 			fmt.Println("Error encoding result:", err)
// 			return
// 		}

// 		fmt.Println(string(output))
// 	}
// }
