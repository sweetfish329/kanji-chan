package handler

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/sweetfish329/kanji-chan/backend/internal/auth"
)

const UserKey = "user"

// AuthMiddleware ログイン済みの幹事/管理者用の認証ミドルウェア
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		var tokenStr string

		// 1. Authorizationヘッダーを確認
		authHeader := c.Request().Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 2. なければCookieを確認
		if tokenStr == "" {
			cookie, err := c.Cookie("session_token")
			if err == nil {
				tokenStr = cookie.Value
			}
		}

		if tokenStr == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
		}

		claims, err := auth.ValidateJWT(tokenStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
		}

		// Echoコンテキストにユーザー情報を格納
		c.Set(UserKey, claims)
		return next(c)
	}
}

// GetUserFromContext コンテキストから認証済みユーザー情報を取得
func GetUserFromContext(c *echo.Context) (*auth.Claims, bool) {
	claims, ok := c.Get(UserKey).(*auth.Claims)
	return claims, ok
}
