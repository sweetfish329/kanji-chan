package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sweetfish329/kanji-chan/backend/internal/auth"
)

type contextKey string

const UserKey contextKey = "user"

// JSONレスポンスの送信ユーティリティ
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// エラーJSONレスポンスの送信ユーティリティ
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// CORSMiddleware CORS対応用のミドルウェア
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// 開発時はフロントエンドからのアクセスを許可
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware ログイン済みの幹事/管理者用の認証ミドルウェア
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenStr string

		// 1. Authorizationヘッダーを確認
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 2. なければCookieを確認
		if tokenStr == "" {
			cookie, err := r.Cookie("session_token")
			if err == nil {
				tokenStr = cookie.Value
			}
		}

		if tokenStr == "" {
			writeError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		claims, err := auth.ValidateJWT(tokenStr)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// コンテキストにユーザー情報を追加
		ctx := context.WithValue(r.Context(), UserKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext コンテキストから認証済みユーザー情報を取得
func GetUserFromContext(r *http.Request) (*auth.Claims, bool) {
	claims, ok := r.Context().Value(UserKey).(*auth.Claims)
	return claims, ok
}
