package auth

import (
	"crypto/subtle"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"github.com/sweetfish329/kanji-chan/backend/internal/model"
)

// SecureCompare は定数時間 (Constant-Time) で2つの文字列を比較し、タイミング攻撃 (Timing Attack) を防止します。
func SecureCompare(input, expected string) bool {
	return subtle.ConstantTimeCompare([]byte(input), []byte(expected)) == 1
}

var (
	jwtSecret []byte
)

// InitAuth 認証関連の初期設定
func InitAuth() {
	clientID := os.Getenv("OAUTH_CLIENT_ID")
	clientSecret := os.Getenv("OAUTH_CLIENT_SECRET")
	redirectURI := os.Getenv("OAUTH_REDIRECT_URI")
	sessionSecret := os.Getenv("JWT_SECRET")
	if sessionSecret == "" {
		sessionSecret = os.Getenv("SESSION_SECRET")
	}

	// 安全設計(Fail-Safe): 環境変数が未設定の場合はフォールバックせず起動時に強制終了する
	if sessionSecret == "" {
		log.Fatalf("Fatal: JWT_SECRET (or SESSION_SECRET) environment variable is not set. Refusing to run with fallback secret.")
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
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	IsAPIKey bool   `json:"is_api_key,omitempty"`
	jwt.RegisteredClaims
}

// GenerateJWT ユーザーIDからJWTを生成
func GenerateJWT(user *model.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	role := user.Role
	if role == "" {
		role = "user"
	}
	claims := &Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Role:     role,
		IsAPIKey: false,
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
