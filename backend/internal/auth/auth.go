package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"github.com/sweetfish329/kanji-chan/backend/internal/model"
)

var (
	jwtSecret []byte
)

// InitAuth 認証関連の初期設定
func InitAuth() {
	clientID := os.Getenv("OAUTH_CLIENT_ID")
	clientSecret := os.Getenv("OAUTH_CLIENT_SECRET")
	redirectURI := os.Getenv("OAUTH_REDIRECT_URI")
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		bytes := make([]byte, 32)
		if _, err := rand.Read(bytes); err == nil {
			sessionSecret = hex.EncodeToString(bytes)
		} else {
			sessionSecret = fmt.Sprintf("kanji-chan-secret-%d", time.Now().UnixNano())
		}
	}
	jwtSecret = []byte(sessionSecret)

	// Gothプロバイダの登録
	goth.UseProviders(
		google.New(clientID, clientSecret, redirectURI, "email", "profile"),
		github.New(clientID, clientSecret, redirectURI, "user:email", "read:user"),
	)
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
