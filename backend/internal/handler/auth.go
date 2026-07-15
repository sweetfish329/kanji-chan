package handler

import (
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/markbates/goth/gothic"
	"github.com/sweetfish329/kanji-chan/backend/internal/auth"
	"github.com/sweetfish329/kanji-chan/backend/internal/database"
	"github.com/sweetfish329/kanji-chan/backend/internal/model"
)

// HandleLogin OAuthログインの開始 (リダイレクト)
func HandleLogin(c *echo.Context) error {
	provider := os.Getenv("OAUTH_PROVIDER")
	if provider == "" {
		provider = "google"
	}

	// gothicがプロバイダを読み込めるようにクエリパラメータに設定する
	q := c.Request().URL.Query()
	q.Set("provider", provider)
	c.Request().URL.RawQuery = q.Encode()

	// gothicによる認証の開始
	gothic.BeginAuthHandler(c.Response(), c.Request())
	return nil
}

// HandleCallback OAuthコールバックの処理
func HandleCallback(c *echo.Context) error {
	provider := os.Getenv("OAUTH_PROVIDER")
	if provider == "" {
		provider = "google"
	}

	q := c.Request().URL.Query()
	q.Set("provider", provider)
	c.Request().URL.RawQuery = q.Encode()

	gothUser, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "OAuth callback failed: "+err.Error())
	}

	// DBにユーザーが存在するか確認、なければ作成、あれば更新
	var user model.User
	result := database.DB.Where("oauth_provider = ? AND oauth_id = ?", gothUser.Provider, gothUser.UserID).First(&user)
	if result.Error != nil {
		// 新規作成
		user = model.User{
			OAuthProvider: gothUser.Provider,
			OAuthID:       gothUser.UserID,
			Email:         gothUser.Email,
			Name:          gothUser.Name,
		}
		if user.Name == "" {
			user.Name = gothUser.NickName
		}
		if user.Email == "" {
			user.Email = gothUser.NickName + "@github.com" // GitHubのフォールバック用など
		}
		if err := database.DB.Create(&user).Error; err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user: "+err.Error())
		}
	} else {
		// 情報更新
		name := gothUser.Name
		if name == "" {
			name = gothUser.NickName
		}
		user.Name = name
		user.Email = gothUser.Email
		database.DB.Save(&user)
	}

	// JWTトークンの生成
	jwtToken, err := auth.GenerateJWT(&user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate session token: "+err.Error())
	}

	// クッキーにセッションを保存
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    jwtToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	// フロントエンドの管理画面へリダイレクト
	return c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173/admin")
}

// HandleLogout ログアウト処理
func HandleLogout(c *echo.Context) error {
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	}
	c.SetCookie(cookie)
	return c.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// HandleMe 現在のユーザー情報を取得
func HandleMe(c *echo.Context) error {
	claims, ok := GetUserFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	var user model.User
	if err := database.DB.First(&user, claims.UserID).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	return c.JSON(http.StatusOK, user)
}

// HandleUpdateAPIKey 管理画面からGemini APIキーを登録・更新
func HandleUpdateAPIKey(c *echo.Context) error {
	claims, ok := GetUserFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	var req struct {
		GeminiAPIKey string `json:"gemini_api_key"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	var user model.User
	if err := database.DB.First(&user, claims.UserID).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	user.GeminiAPIKey = req.GeminiAPIKey
	if err := database.DB.Save(&user).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update API key")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "API key updated successfully"})
}
