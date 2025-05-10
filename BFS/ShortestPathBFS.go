package bfs

import (
	"math"
	"stima-2-be/Element"
)

func FindShortestRecipeBFS(target string) Element.Tree {
	visited := make(map[string]bool)
	queue := Element.GetElements(target)

	var bestTree Element.Tree
	bestTier := math.MaxInt32

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.Root] {
			continue
		}
		visited[current.Root] = true

		tree := Element.BuildTreeWrapper(current)
		if Element.ValidateTree(tree) {
			tier := tree.GetTier()
			if tier < bestTier {
				bestTier = tier
				bestTree = tree
			}
		}
	}

	return bestTree
}
