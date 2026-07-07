package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sweetfish329/kanji-chan/backend/internal/auth"
	"github.com/sweetfish329/kanji-chan/backend/internal/database"
	"github.com/sweetfish329/kanji-chan/backend/internal/handler"
)

func main() {
	// ローカル開発時は .env ファイルをロード (コンテナ時は環境変数が直接渡されるため無視してOK)
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// データベース接続の初期化
	_, err := database.InitDB()
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	// 認証の初期設定
	auth.InitAuth()

	mux := http.NewServeMux()

	// 共通・認証 (パブリック)
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok", "message":"Kanji-Chan API is running"}`))
	})
	mux.HandleFunc("GET /api/auth/login", handler.HandleLogin)
	mux.HandleFunc("GET /api/auth/callback", handler.HandleCallback)
	mux.HandleFunc("POST /api/auth/logout", handler.HandleLogout)

	// 認証保護されるAPI用サブマルチプレクサ
	authMux := http.NewServeMux()
	authMux.HandleFunc("GET /api/auth/me", handler.HandleMe)
	authMux.HandleFunc("POST /api/auth/apikey", handler.HandleUpdateAPIKey)
	authMux.HandleFunc("POST /api/events", handler.HandleCreateEvent)
	authMux.HandleFunc("GET /api/events", handler.HandleListEvents)
	authMux.HandleFunc("PUT /api/events/{id}", handler.HandleUpdateEvent)
	authMux.HandleFunc("DELETE /api/events/{id}", handler.HandleDeleteEvent)

	// イベント詳細・回答登録 (パブリック)
	mux.HandleFunc("GET /api/events/{id}", handler.HandleGetEvent)
	mux.HandleFunc("POST /api/events/{id}/responses", handler.HandleAddResponse)
	mux.HandleFunc("DELETE /api/events/{id}/responses/{response_id}", handler.HandleDeleteResponse)

	// ルーター統合
	// 認証ミドルウェアで保護されたパスをメインの ServeMux に登録
	protectedHandler := handler.AuthMiddleware(authMux)
	
	// パスマッチングルールのため、認証が必要なエンドポイントは authMux を通す
	// 簡易的に ServeMux 内でパス単位でミドルウェアを当てるか、
	// メインの ServeMux にフォールバックとしてカスタムハンドラーを当てる
	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 特定の認証が必要なパスのみミドルウェアを通す
		if r.URL.Path == "/api/auth/me" || 
		   r.URL.Path == "/api/auth/apikey" || 
		   (r.URL.Path == "/api/events" && (r.Method == "POST" || r.Method == "GET")) ||
		   (strings.HasPrefix(r.URL.Path, "/api/events/") && !strings.HasSuffix(r.URL.Path, "/responses") && !strings.Contains(r.URL.Path, "/responses/") && (r.Method == "PUT" || r.Method == "DELETE")) {
			protectedHandler.ServeHTTP(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	})

	// CORSミドルウェアを適用して起動
	corsHandler := handler.CORSMiddleware(mainHandler)

	log.Printf("Starting Kanji-Chan backend server on port %s...", port)
	if err := http.ListenAndServe(":"+port, corsHandler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
// ※ strings パッケージのインポートが必要です。追加で修正します。
