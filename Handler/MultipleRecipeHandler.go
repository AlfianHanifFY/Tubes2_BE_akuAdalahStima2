package handler

import (
	"net/http"
)

func MultipleRecipeHandler(w http.ResponseWriter, r *http.Request) {
	// name := r.URL.Query().Get("element")
	// tipe := r.URL.Query().Get("type")
	// countStr := r.URL.Query().Get("count")

	// count, err := strconv.Atoi(countStr)
	// if err != nil {
	// 	count = 0 // default value
	// }

	// switch tipe {
	// case "BFS":
	// 	result := bfs.MultipleRecipe(name, count)
	// case "DFS":
	// 	result := dfs.MultipleRecipe(name, count)
	// default:
	// 	http.Error(w, "Invalid type parameter: must be 'BFS' or 'DFS'", http.StatusBadRequest)
	// 	return
	// }

	// w.Header().Set("Content-Type", "application/json")
	// if err := json.NewEncoder(w).Encode(result); err != nil {
	// 	http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	// 	fmt.Println("JSON encode error:", err)
	// }
}
