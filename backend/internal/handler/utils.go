package handler

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/sweetfish329/kanji-chan/backend/internal/auth"
)

const UserKey = "user"

// AuthMiddleware ログイン済みの幹事/管理者用の認証ミドルウェア (JWTおよびAPIキーに対応)
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		var claims *auth.Claims
		var err error

		// 1. Authorization ヘッダーを確認
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader != "" {
			tokenStr := authHeader
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			}

			// API キーの場合 (kc_ プレフィックス)
			if strings.HasPrefix(tokenStr, auth.APIKeyPrefix) {
				claims, err = auth.ValidateAPIKey(tokenStr)
			} else {
				claims, err = auth.ValidateJWT(tokenStr)
			}
		}

		// 2. X-API-Key ヘッダーを確認
		if claims == nil {
			apiKeyHeader := c.Request().Header.Get("X-API-Key")
			if apiKeyHeader != "" {
				claims, err = auth.ValidateAPIKey(apiKeyHeader)
			}
		}

		// 3. クエリパラメーター api_key を確認
		if claims == nil {
			queryApiKey := c.QueryParam("api_key")
			if queryApiKey != "" {
				claims, err = auth.ValidateAPIKey(queryApiKey)
			}
		}

		// 4. Cookie (session_token) を確認
		if claims == nil {
			cookie, cookieErr := c.Cookie("session_token")
			if cookieErr == nil && cookie.Value != "" {
				claims, err = auth.ValidateJWT(cookie.Value)
			}
		}

		if claims == nil || err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required (Invalid or missing token/API key)")
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
