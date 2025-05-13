package Element

import (
	"fmt"
	"strconv"
)

// Struct pohon berbasis elemenU
type Tree struct {
	Root     Element
	Children []Tree
}

// Inisialisasi node pohon dari elemen
func CreateTree(e Element) Tree {
	var t Tree
	t.Root = e
	return t
}

// Tambah child ke tree
func (t *Tree) AddChild(child Tree) {
	t.Children = append(t.Children, child)
}

// Ambil nama root dari tree
func (t Tree) GetRootName() string {
	return t.Root.Root
}

func (t Tree) GetChildren() []Tree {
	return t.Children
}

func (t Tree) GetTier() int {
	i, _ := strconv.Atoi(t.Root.Tier)
	return i
}

// ini buat nge tes api endpoint aja
func BuildTree(e Element, visited map[string]bool, usedRoots map[string]bool, depth int) Tree {
	if depth > 100 {
		return Tree{}
	}

	if visited[e.Root] || usedRoots[e.Root] {
		return Tree{}
	}

	// Buat salinan untuk path saat ini (hindari siklus)
	pathVisited := make(map[string]bool)
	for k, v := range visited {
		pathVisited[k] = v
	}
	pathVisited[e.Root] = true

	// Tandai root ini sebagai sudah digunakan
	usedRoots[e.Root] = true

	tree := CreateTree(e)

	if tree.GetTier() != 0 {
		// === Anak Kiri ===
		leftChildren := e.LeftChildren()
		if len(leftChildren) > 0 {
			var best Tree
			bestTier := 999999

			for _, left := range leftChildren {
				if usedRoots[left.Root] {
					continue
				}
				leftTree := BuildTree(left, pathVisited, usedRoots, depth+1)
				if leftTree.Root.Root != "" && leftTree.GetTier() < bestTier {
					best = leftTree
					bestTier = leftTree.GetTier()
				}
			}

			if best.Root.Root != "" {
				tree.AddChild(best)
			}
		}

		// === Anak Kanan ===
		rightChildren := e.RightChildren()
		if len(rightChildren) > 0 {
			var best Tree
			bestTier := 999999

			for _, right := range rightChildren {
				if usedRoots[right.Root] {
					continue
				}
				rightTree := BuildTree(right, pathVisited, usedRoots, depth+1)
				if rightTree.Root.Root != "" && rightTree.GetTier() < bestTier {
					best = rightTree
					bestTier = rightTree.GetTier()
				}
			}

			if best.Root.Root != "" {
				tree.AddChild(best)
			}
		}
	}

	return tree
}

func BuildTreeWrapper(e Element) Tree {
	visited := make(map[string]bool)
	usedRoots := make(map[string]bool)
	return BuildTree(e, visited, usedRoots, 0)
}

func TreesEqual(a, b Tree) bool {
	if a.Root.Root != b.Root.Root ||
		a.Root.Left != b.Root.Left ||
		a.Root.Right != b.Root.Right ||
		a.Root.Tier != b.Root.Tier {
		return false
	}

	if len(a.Children) != len(b.Children) {
		return false
	}

	for i := range a.Children {
		if !TreesEqual(a.Children[i], b.Children[i]) {
			return false
		}
	}
	return true
}

func AreAllTreesUnique(trees []Tree) bool {
	for i := 0; i < len(trees); i++ {
		for j := i + 1; j < len(trees); j++ {
			if TreesEqual(trees[i], trees[j]) {
				fmt.Println("not unique")
				return false
			}
		}
	}
	fmt.Println("unique recipe")
	return true
}
