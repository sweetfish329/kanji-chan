package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sweetfish329/kanji-chan/backend/internal/auth"
	"github.com/sweetfish329/kanji-chan/backend/internal/database"
	"github.com/sweetfish329/kanji-chan/backend/internal/model"
)

// HandleLogin OAuthログインの開始 (リダイレクト)
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	// CSRF防止用のstate値を生成
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	// 本来はstateをセッション等に保存して検証すべきだが、デモ簡略化のため一旦Cookieに保持
	cookie := &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, cookie)

	config := auth.GetOAuthConfig()
	url := config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleCallback OAuthコールバックの処理
func HandleCallback(w http.ResponseWriter, r *http.Request) {
	// stateの検証
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || r.FormValue("state") != stateCookie.Value {
		writeError(w, http.StatusBadRequest, "Invalid OAuth state")
		return
	}

	code := r.FormValue("code")
	config := auth.GetOAuthConfig()
	token, err := config.Exchange(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to exchange token: "+err.Error())
		return
	}

	// ユーザー情報の取得
	oauthUser, err := auth.GetUserInfo(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get user info: "+err.Error())
		return
	}

	// DBにユーザーが存在するか確認、なければ作成、あれば更新
	var user model.User
	result := database.DB.Where("oauth_provider = ? AND oauth_id = ?", oauthUser.OAuthProvider, oauthUser.OAuthID).First(&user)
	if result.Error != nil {
		// 新規作成
		user = *oauthUser
		if err := database.DB.Create(&user).Error; err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create user: "+err.Error())
			return
		}
	} else {
		// 情報更新
		user.Name = oauthUser.Name
		user.Email = oauthUser.Email
		database.DB.Save(&user)
	}

	// JWTトークンの生成
	jwtToken, err := auth.GenerateJWT(&user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate session token: "+err.Error())
		return
	}

	// クッキーにセッションを保存
	sessionCookie := &http.Cookie{
		Name:     "session_token",
		Value:    jwtToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode, // SPA連携のため
	}
	http.SetCookie(w, sessionCookie)

	// フロントエンドの管理画面などへリダイレクト
	// 開発用フロントエンドのアドレス (環境変数や設定から取得すべきだが、デモ用は http://localhost:5173/admin とする)
	http.Redirect(w, r, "http://localhost:5173/admin", http.StatusTemporaryRedirect)
}

// HandleLogout ログアウト処理
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	sessionCookie := &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, sessionCookie)
	writeJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// HandleMe 現在のユーザー情報を取得
func HandleMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetUserFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var user model.User
	if err := database.DB.First(&user, claims.UserID).Error; err != nil {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// HandleUpdateAPIKey 管理画面からGemini APIキーを登録・更新
func HandleUpdateAPIKey(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetUserFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		GeminiAPIKey string `json:"gemini_api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var user model.User
	if err := database.DB.First(&user, claims.UserID).Error; err != nil {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	user.GeminiAPIKey = req.GeminiAPIKey
	if err := database.DB.Save(&user).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update API key")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "API key updated successfully"})
}
