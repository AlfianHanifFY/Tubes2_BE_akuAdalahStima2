package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	bfs "stima-2-be/BFS"
	"stima-2-be/Element"
	"strconv"
	"strings"
)

func BFSHandler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get and clean parameters
	name := strings.TrimSpace(r.URL.Query().Get("element"))
	countStr := r.URL.Query().Get("count")

	// Default count if not provided
	count := 10 // Default value
	if countStr != "" {
		var err error
		count, err = strconv.Atoi(countStr)
		if err != nil {
			fmt.Println("Conversion error:", err)
			http.Error(w, "Invalid count parameter", http.StatusBadRequest)
			return
		}
	}

	fmt.Printf("Processing BFS request for element: %s, count: %d\n", name, count)

	// Get recipe map first (same as in DFS implementation)
	recipeMap := Element.BuildRecipeMap(Element.GetAllElement())

	// Get recipes using BFS algorithm with multithreading
	// Using the new function that returns both trees and metrics
	result, info := bfs.MultipleRecipe(name, recipeMap, count)

	fmt.Printf("Debug: BFS result count=%d\n", len(result))

	// Format the response to match DFS handler
	response := []interface{}{info, result}

	// Set content type and send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		fmt.Println("JSON encode error:", err)
	}
}
