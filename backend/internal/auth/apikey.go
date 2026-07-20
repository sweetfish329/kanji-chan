package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/sweetfish329/kanji-chan/backend/internal/database"
	"github.com/sweetfish329/kanji-chan/backend/internal/model"
)

const APIKeyPrefix = "kc_"

// GenerateAPIKey 新しい API キーを生成してデータベースに保存
func GenerateAPIKey(name string, userID uint) (string, *model.ApiKey, error) {
	if name == "" {
		name = "デフォルト API キー"
	}

	// 32バイトの暗号学的に安全な乱数を生成
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	rawKey := APIKeyPrefix + hex.EncodeToString(bytes) // 例: kc_6f... (67文字)

	// SHA-256 ハッシュを算出
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	// UI表示用のプレフィックス (例: "kc_6f3a9b...")
	keyPrefix := rawKey[:9] + "..."

	role := "admin"
	var user model.User
	if err := database.DB.First(&user, userID).Error; err == nil && user.Role != "" {
		role = user.Role
	}

	apiKey := &model.ApiKey{
		UserID:    userID,
		Name:      name,
		Role:      role,
		KeyPrefix: keyPrefix,
		KeyHash:   keyHash,
		CreatedAt: time.Now(),
	}

	if err := database.DB.Create(apiKey).Error; err != nil {
		return "", nil, fmt.Errorf("failed to save API key to database: %w", err)
	}

	return rawKey, apiKey, nil
}

// ValidateAPIKey 入力された API キーの妥当性を検証し、ユーザー情報 (Claims) を返す
func ValidateAPIKey(rawKey string) (*Claims, error) {
	if rawKey == "" {
		return nil, errors.New("API key is empty")
	}

	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	var apiKey model.ApiKey
	if err := database.DB.Where("key_hash = ?", keyHash).First(&apiKey).Error; err != nil {
		return nil, errors.New("invalid or expired API key")
	}

	var user model.User
	if err := database.DB.First(&user, apiKey.UserID).Error; err != nil {
		return nil, errors.New("user associated with API key not found")
	}

	// 最終利用日時を更新 (非同期/Goroutine で行いレスポンスに遅延を出さない)
	go func(id uint) {
		now := time.Now()
		database.DB.Model(&model.ApiKey{}).Where("id = ?", id).Update("last_used_at", now)
	}(apiKey.ID)

	role := apiKey.Role
	if role == "" {
		role = user.Role
	}
	if role == "" {
		role = "admin"
	}

	claims := &Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Role:     role,
		IsAPIKey: true,
	}

	return claims, nil
}
