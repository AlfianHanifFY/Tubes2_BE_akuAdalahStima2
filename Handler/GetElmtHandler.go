package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"stima-2-be/Element"
)

func GetElmtHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("element")
	if name == "" {
		http.Error(w, "Missing 'element' parameter", http.StatusBadRequest)
		return
	}

	response := Element.GetElements(name)
	fmt.Println("sukses")
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		fmt.Println("JSON encode error:", err)
	}
}
