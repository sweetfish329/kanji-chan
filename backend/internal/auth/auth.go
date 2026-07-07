package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sweetfish329/kanji-chan/backend/internal/model"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

var (
	oauthConfig *oauth2.Config
	jwtSecret   []byte
)

// InitAuth 認証関連の初期設定
func InitAuth() {
	provider := os.Getenv("OAUTH_PROVIDER")
	clientID := os.Getenv("OAUTH_CLIENT_ID")
	clientSecret := os.Getenv("OAUTH_CLIENT_SECRET")
	redirectURI := os.Getenv("OAUTH_REDIRECT_URI")
	sessionSecret := os.Getenv("SESSION_SECRET")

	if sessionSecret == "" {
		sessionSecret = "kanji-chan-default-secret-key"
	}
	jwtSecret = []byte(sessionSecret)

	var endpoint oauth2.Endpoint
	var scopes []string

	if provider == "github" {
		endpoint = github.Endpoint
		scopes = []string{"user:email", "read:user"}
	} else {
		// デフォルトはGoogle
		endpoint = google.Endpoint
		scopes = []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		}
	}

	oauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Endpoint:     endpoint,
		Scopes:       scopes,
	}
}

// GetOAuthConfig OAuth2設定の取得
func GetOAuthConfig() *oauth2.Config {
	return oauthConfig
}

// Claims JWTのクレーム構造体
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

// GenerateJWT ユーザーIDからJWTを生成
func GenerateJWT(user *model.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateJWT JWTトークンの検証
func ValidateJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// GoogleUser Googleのユーザー情報レスポンス
type GoogleUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// GitHubUser GitHubのユーザー情報レスポンス
type GitHubUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Login string `json:"login"`
}

// GetUserInfo OAuthプロバイダからユーザー情報を取得
func GetUserInfo(ctx context.Context, token *oauth2.Token) (*model.User, error) {
	provider := os.Getenv("OAUTH_PROVIDER")
	client := oauthConfig.Client(ctx, token)

	if provider == "github" {
		req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var ghUser GitHubUser
		if err := json.NewDecoder(resp.Body).Decode(&ghUser); err != nil {
			return nil, err
		}

		name := ghUser.Name
		if name == "" {
			name = ghUser.Login
		}

		// メールアドレスの取得 (GitHubは公開設定によってはemailが空になるため、emails APIを別途叩くか、フォールバック)
		email := ghUser.Email
		if email == "" {
			email = fmt.Sprintf("%s@github.com", ghUser.Login)
		}

		return &model.User{
			OAuthProvider: "github",
			OAuthID:       fmt.Sprintf("%d", ghUser.ID),
			Email:         email,
			Name:          name,
		}, nil
	} else {
		// Google
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var gUser GoogleUser
		if err := json.NewDecoder(resp.Body).Decode(&gUser); err != nil {
			return nil, err
		}

		return &model.User{
			OAuthProvider: "google",
			OAuthID:       gUser.ID,
			Email:         gUser.Email,
			Name:          gUser.Name,
		}, nil
	}
}
