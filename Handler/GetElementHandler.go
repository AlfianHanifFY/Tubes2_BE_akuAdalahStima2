package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"stima-2-be/Element"
)

func GetElementHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("element")
	result := Element.GetElements(name)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		fmt.Println("JSON encode error:", err)
	}
}
