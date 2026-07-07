package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// ヘルスチェック用エンドポイント
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok", "message":"Kanji-Chan API is running"}`))
	})

	log.Printf("Starting Kanji-Chan backend server on port %s...", port)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
// ※ r *http.Type はタイポで、正しくは r *http.Request です。修正します。
