package main

import (
	"fmt"
	"net/http"
	handler "stima-2-be/Handler"
)

// CORS middleware untuk menambahkan header Access-Control-Allow-Origin
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Atau bisa menggunakan "http://localhost:3000" untuk lebih spesifik
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Jika request method OPTIONS, langsung beri respons OK (pre-flight)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Lanjutkan ke handler asli
		next.ServeHTTP(w, r)
	}
}

func main() {
	// Menambahkan CORS middleware ke handler

	http.HandleFunc("/", enableCORS(handler.ScrapHandler))
	http.HandleFunc("/ShortestPath", enableCORS(handler.ShortestPathHandler))
	http.HandleFunc("/TestTree", enableCORS(handler.TestTreeHandler))
	http.HandleFunc("/BFS", enableCORS(handler.BFSHandler))
	http.HandleFunc("/DFS", enableCORS(handler.DFSHandler))
	http.HandleFunc("/MultipleRecipeBFS", enableCORS(handler.MultipleRecipeHandlerBFS))

	// Jalankan server pada port 8080
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
