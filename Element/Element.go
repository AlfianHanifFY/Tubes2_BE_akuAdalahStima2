package Element

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

// Struct dasar Element
type Element struct {
	Root  string `json:"root"`
	Left  string `json:"Left"`
	Right string `json:"Right"`
	Tier  string `json:"Tier"`
}

// Menyimpan semua elemen dari file agar tidak perlu load berulang
var allElements []Element

func GetAllElement() []Element {
	return allElements
}

// Load dari file JSON hanya sekali
func LoadElementsFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &allElements)
	if err != nil {
		return err
	}
	return nil
}

// Ambil semua elemen dengan root tertentu (case-insensitive)
func GetElements(rootName string) []Element {
	var result []Element
	for _, elem := range allElements {
		if strings.EqualFold(elem.Root, rootName) {
			result = append(result, elem)
		}
	}

	// Debugging log
	if len(result) == 0 {
		log.Printf("No elements found with root name: %s", rootName)
	} else {
		log.Printf("Found %d elements for root name: %s", len(result), rootName)
	}

	return result
}

// Metode untuk ambil child dari suatu elemen
func (e Element) LeftChildren() []Element {
	return GetElements(e.Left)
}

func (e Element) RightChildren() []Element {
	return GetElements(e.Right)
}

// Define base components, jika 'time' ga masuk
var BaseComponents = map[string]bool{
	"air":   true,
	"earth": true,
	"fire":  true,
	"water": true,
	"time":  true,
}

// Kelompokin berdasarkan rootnya
func BuildRecipeMap(recipes []Element) map[string][]Element {
	recipeMap := make(map[string][]Element)
	for _, r := range recipes {
		recipeMap[strings.ToLower(r.Root)] = append(recipeMap[strings.ToLower(r.Root)], r)
	}
	return recipeMap
}

// Check if an item is a base component
func IsBaseComponent(item string) bool {
	return BaseComponents[strings.ToLower(item)]
}

// Check if all leaf nodes are base components
func ValidateTree(tree Tree) bool {
	// If this is a leaf node
	if len(tree.Children) == 0 {
		return IsBaseComponent(tree.Root.Root)
	}

	// Check all children
	for _, child := range tree.Children {
		if !ValidateTree(child) {
			return false
		}
	}

	return true
}

func ParseTier(tierStr string) int {
	var tier int
	fmt.Sscanf(tierStr, "%d", &tier)
	return tier
}
