package bfs

import (
	"fmt"
	"math"
	"stima-2-be/Element"
	"strings"
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

// Function untuk mencari semua resep yang mungkin dari BFS
func findRecipesBFS(root string, recipeMap map[string][]Element.Element, tierLimit int, limit int) ([]Element.Element, int64) {
	var nodesVisited int64 = 0
	var recipes []Element.Element

	// Jika root adalah komponen dasar, return kosong
	if Element.IsBaseComponent(root) || root == "time" {
		return recipes, nodesVisited
	}

	// Inisialisasi queue dan visited
	queue := &Queue{}
	visited := make(map[string]bool)

	// Mulai dari root
	queue.Enqueue(root)

	// BFS traversal
	for !queue.IsEmpty() && len(recipes) < limit {
		current := queue.Dequeue()
		current = strings.ToLower(current)

		nodesVisited++
		fmt.Printf("[DEBUG] Visiting: %s (Nodes Visited: %d)\n", current, nodesVisited)

		// Skip jika sudah dikunjungi
		if visited[current] {
			continue
		}

		// Tandai sebagai sudah dikunjungi
		visited[current] = true

		// Cari resep yang cocok
		currentRecipes, exists := recipeMap[current]
		if !exists {
			continue
		}

		// Tambahkan resep ke hasil
		for _, recipe := range currentRecipes {
			recipeInt := Element.ParseTier(recipe.Tier)
			if recipeInt < tierLimit {
				recipes = append(recipes, recipe)

				// Tambahkan child nodes ke queue
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

		// Batasi jumlah resep
		if len(recipes) >= limit {
			recipes = recipes[:limit]
			break
		}
	}

	return recipes, nodesVisited
}

// Function untuk membangun semua kemungkinan tree dari resep
func buildAllTreesFromRecipe(recipe Element.Element, recipeMap map[string][]Element.Element, visited map[string]bool, tierLimit int, limit int, nodesVisited *int64) []Element.Tree {
	*nodesVisited++

	// Cek apakah sudah pernah dikunjungi untuk menghindari cycle
	newVisited := cloneMap(visited)
	newVisited[strings.ToLower(recipe.Root)] = true

	// Siapkan array untuk menyimpan semua kemungkinan tree
	var resultTrees []Element.Tree

	// Build left subtrees
	left := strings.ToLower(recipe.Left)
	var leftTrees []Element.Tree

	if Element.IsBaseComponent(left) {
		*nodesVisited++
		// Jika komponen dasar, hanya ada satu kemungkinan
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
			// Cek semua kemungkinan resep untuk left
			for _, leftRecipe := range leftRecipes {
				leftTierInt := Element.ParseTier(leftRecipe.Tier)
				if leftTierInt < tierLimit {
					// Hitung berapa banyak pohon yang bisa dialokasikan untuk subtree ini
					subLimit := limit
					if subLimit > 10 {
						subLimit = 10 // Batasi untuk menghindari eksplorasi yang terlalu luas
					}

					leftSubtrees := buildAllTreesFromRecipe(leftRecipe, recipeMap, newVisited, leftTierInt, subLimit, nodesVisited)
					leftTrees = append(leftTrees, leftSubtrees...)

					// Jika sudah cukup resep ditemukan, hentikan
					if len(leftTrees) >= subLimit {
						break
					}
				}
			}
		}
	}

	// Jika tidak ada leftTrees yang valid, return kosong
	if len(leftTrees) == 0 {
		return []Element.Tree{}
	}

	// Build right subtrees
	right := strings.ToLower(recipe.Right)
	var rightTrees []Element.Tree

	if Element.IsBaseComponent(right) {
		*nodesVisited++
		// Jika komponen dasar, hanya ada satu kemungkinan
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
			// Cek semua kemungkinan resep untuk right
			for _, rightRecipe := range rightRecipes {
				rightTierInt := Element.ParseTier(rightRecipe.Tier)
				if rightTierInt < tierLimit {
					// Hitung berapa banyak pohon yang bisa dialokasikan untuk subtree ini
					subLimit := limit
					if subLimit > 10 {
						subLimit = 10 // Batasi untuk menghindari eksplorasi yang terlalu luas
					}

					rightSubtrees := buildAllTreesFromRecipe(rightRecipe, recipeMap, newVisited, rightTierInt, subLimit, nodesVisited)
					rightTrees = append(rightTrees, rightSubtrees...)

					// Jika sudah cukup resep ditemukan, hentikan
					if len(rightTrees) >= subLimit {
						break
					}
				}
			}
		}
	}

	// Jika tidak ada rightTrees yang valid, return kosong
	if len(rightTrees) == 0 {
		return []Element.Tree{}
	}

	// Kombinasikan leftTrees dan rightTrees untuk membuat semua kemungkinan tree
	for _, lt := range leftTrees {
		for _, rt := range rightTrees {
			tree := Element.Tree{
				Root:     recipe,
				Children: []Element.Tree{lt, rt},
			}
			resultTrees = append(resultTrees, tree)

			// Batasi jumlah tree yang dibuat
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

	// Gunakan BFS untuk menemukan semua resep yang mungkin
	recipes, visitedCount := findRecipesBFS(root, recipeMap, math.MaxInt32, limit*2)
	nodesVisited += visitedCount

	// Bangun tree untuk setiap resep root yang ditemukan
	for _, recipe := range recipes {
		// Hanya gunakan resep dengan root yang sesuai
		if strings.ToLower(recipe.Root) == strings.ToLower(root) {
			visited := make(map[string]bool)
			tierInt := Element.ParseTier(recipe.Tier)

			// Build semua kemungkinan trees untuk resep ini
			trees := buildAllTreesFromRecipe(recipe, recipeMap, visited, tierInt, limit-len(resultTrees), &nodesVisited)
			resultTrees = append(resultTrees, trees...)

			// Batasi jumlah tree
			if len(resultTrees) >= limit {
				resultTrees = resultTrees[:limit]
				break
			}
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
