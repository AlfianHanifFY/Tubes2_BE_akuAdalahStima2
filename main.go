package main

import (
	"fmt"
	"net/http"
	handler "stima-2-be/Handler"
)

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func main() {

	http.HandleFunc("/Scrap", enableCORS(handler.ScrapHandler))
	http.HandleFunc("/BFS", enableCORS(handler.BFSHandler))
	http.HandleFunc("/DFS", enableCORS(handler.DFSHandler))

	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
