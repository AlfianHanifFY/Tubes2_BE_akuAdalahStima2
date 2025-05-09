package Element

import (
	"encoding/json"
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
