package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	dfs "stima-2-be/DFS"
	"stima-2-be/Element"
	"strconv"
)

func DFSHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("element")
	countStr := r.URL.Query().Get("count")
	count, err := strconv.Atoi(countStr)
	if err != nil {
		fmt.Println("Conversion error:", err)
	} else {
		fmt.Println("Converted int:", count)
	}
	recipeMap := Element.BuildRecipeMap(Element.GetAllElement())
	result := dfs.MultipleRecipe(name, recipeMap, count)

	fmt.Print(result)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		fmt.Println("JSON encode error:", err)
	}
}
