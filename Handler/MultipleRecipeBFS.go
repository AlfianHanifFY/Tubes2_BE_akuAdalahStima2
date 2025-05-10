package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	bfs "stima-2-be/BFS"
)

func MultipleRecipeHandlerBFS(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("element")
	countStr := r.URL.Query().Get("count")
	count, err := strconv.Atoi(countStr)
	if err != nil {
		fmt.Println("Conversion error:", err)
	} else {
		fmt.Println("Converted int:", count)
	}

	result := bfs.MultipleRecipesBFS(name, count)

	fmt.Print(result)
	fmt.Printf("Debug: BFS result=%v\n", result)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		fmt.Println("JSON encode error:", err)
	}
}