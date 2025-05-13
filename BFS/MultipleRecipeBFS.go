package bfs

import (
	"fmt"
	"math"
	"stima-2-be/Element"
	"strings"
	"sync"
	"time"
)

type MetricsResult struct {
	NodesVisited  int64  `json:"nodes_visited"`
	Duration      int64  `json:"duration_ms"`
	DurationHuman string `json:"duration_human"`
}

// Queue untuk BFS
type Queue struct {
	items []string
}

func (q *Queue) Enqueue(item string) {
	q.items = append(q.items, item)
}

func (q *Queue) Dequeue() string {
	if len(q.items) == 0 {
		return ""
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q *Queue) IsEmpty() bool {
	return len(q.items) == 0
}

func cloneMap(original map[string]bool) map[string]bool {
	copy := make(map[string]bool)
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

// nyari semua resep make BFS
func findRecipesBFS(root string, recipeMap map[string][]Element.Element, tierLimit int, limit int) ([]Element.Element, int64) {
	var nodesVisited int64 = 0
	var recipes []Element.Element

	// Jika root adalah komponen dasar, return kosong
	if Element.IsBaseComponent(root) || root == "time" {
		return recipes, nodesVisited
	}

	queue := &Queue{}
	visited := make(map[string]bool)
	queue.Enqueue(root)

	for !queue.IsEmpty() && len(recipes) < limit {
		current := queue.Dequeue()
		current = strings.ToLower(current)
		nodesVisited++

		// Skip klo udah visited
		if visited[current] {
			continue
		}
		visited[current] = true

		currentRecipes, exists := recipeMap[current]
		if !exists {
			continue
		}

		for _, recipe := range currentRecipes {
			recipeInt := Element.ParseTier(recipe.Tier)
			if recipeInt < tierLimit {
				recipes = append(recipes, recipe)

				left := strings.ToLower(recipe.Left)
				right := strings.ToLower(recipe.Right)

				if !visited[left] {
					queue.Enqueue(left)
				}
				if !visited[right] {
					queue.Enqueue(right)
				}
			}
		}

		if len(recipes) >= limit {
			recipes = recipes[:limit]
			break
		}
	}

	return recipes, nodesVisited
}

func buildAllTreesFromRecipe(recipe Element.Element, recipeMap map[string][]Element.Element, visited map[string]bool, tierLimit int, limit int, nodesVisited *int64) []Element.Tree {
	*nodesVisited++

	newVisited := cloneMap(visited)
	newVisited[strings.ToLower(recipe.Root)] = true

	var resultTrees []Element.Tree

	left := strings.ToLower(recipe.Left)
	var leftTrees []Element.Tree

	if Element.IsBaseComponent(left) {
		*nodesVisited++
		leftTrees = append(leftTrees, Element.Tree{
			Root: Element.Element{
				Root:  left,
				Left:  "",
				Right: "",
				Tier:  "0",
			},
			Children: nil,
		})
	} else if !newVisited[left] {
		leftRecipes, exists := recipeMap[left]
		if exists {
			for _, leftRecipe := range leftRecipes {
				leftTierInt := Element.ParseTier(leftRecipe.Tier)
				if leftTierInt < tierLimit {
					subLimit := limit

					leftSubtrees := buildAllTreesFromRecipe(leftRecipe, recipeMap, newVisited, leftTierInt, subLimit, nodesVisited)
					leftTrees = append(leftTrees, leftSubtrees...)
					if len(leftTrees) >= subLimit {
						break
					}
				}
			}
		}
	}

	if len(leftTrees) == 0 {
		return []Element.Tree{}
	}

	right := strings.ToLower(recipe.Right)
	var rightTrees []Element.Tree

	if Element.IsBaseComponent(right) {
		*nodesVisited++
		rightTrees = append(rightTrees, Element.Tree{
			Root: Element.Element{
				Root:  right,
				Left:  "",
				Right: "",
				Tier:  "0",
			},
			Children: nil,
		})
	} else if !newVisited[right] {
		rightRecipes, exists := recipeMap[right]
		if exists {
			for _, rightRecipe := range rightRecipes {
				rightTierInt := Element.ParseTier(rightRecipe.Tier)
				if rightTierInt < tierLimit {
					subLimit := limit
					if subLimit > 10 {
						subLimit = 10
					}
					rightSubtrees := buildAllTreesFromRecipe(rightRecipe, recipeMap, newVisited, rightTierInt, subLimit, nodesVisited)
					rightTrees = append(rightTrees, rightSubtrees...)
					if len(rightTrees) >= subLimit {
						break
					}
				}
			}
		}
	}

	if len(rightTrees) == 0 {
		return []Element.Tree{}
	}

	for _, lt := range leftTrees {
		for _, rt := range rightTrees {
			tree := Element.Tree{
				Root:     recipe,
				Children: []Element.Tree{lt, rt},
			}
			resultTrees = append(resultTrees, tree)
			if len(resultTrees) >= limit {
				return resultTrees
			}
		}
	}

	return resultTrees
}

func buildTreesBFS(root string, recipeMap map[string][]Element.Element, limit int) ([]Element.Tree, int64) {
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
		}, 1
	}

	var nodesVisited int64 = 0
	var resultTrees []Element.Tree

	recipes, visitedCount := findRecipesBFS(root, recipeMap, math.MaxInt32, limit*2)
	nodesVisited += visitedCount

	var mu sync.Mutex
	var wg sync.WaitGroup

	treeChan := make(chan []Element.Tree, len(recipes))

	for _, recipe := range recipes {
		if strings.ToLower(recipe.Root) != strings.ToLower(root) {
			continue
		}

		wg.Add(1)
		go func(r Element.Element) {
			defer wg.Done()

			visited := make(map[string]bool)
			tierInt := Element.ParseTier(r.Tier)
			var localVisited int64 = 0

			trees := buildAllTreesFromRecipe(r, recipeMap, visited, tierInt, limit, &localVisited)

			mu.Lock()
			nodesVisited += localVisited
			mu.Unlock()

			treeChan <- trees
		}(recipe)
	}

	wg.Wait()
	close(treeChan)

	for trees := range treeChan {
		for _, tree := range trees {
			if len(resultTrees) >= limit {
				break
			}
			resultTrees = append(resultTrees, tree)
		}
		if len(resultTrees) >= limit {
			break
		}
	}

	return resultTrees, nodesVisited
}

func MultipleRecipe(name string, recipeMap map[string][]Element.Element, count int) ([]Element.Tree, MetricsResult) {
	startTime := time.Now()

	name = strings.ToLower(name)
	var trees []Element.Tree
	var nodesVisited int64

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
		nodesVisited = 1
	} else {
		trees, nodesVisited = buildTreesBFS(name, recipeMap, count)
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

func PrintTree(t Element.Tree, indent string) {
	fmt.Printf("%s%s (Tier: %s)\n", indent, t.Root.Root, t.Root.Tier)
	if len(t.Children) > 0 {
		for _, child := range t.Children {
			PrintTree(child, indent+"  ")
		}
	}
}
