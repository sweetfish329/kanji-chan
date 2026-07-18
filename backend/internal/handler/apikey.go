package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/sweetfish329/kanji-chan/backend/internal/auth"
	"github.com/sweetfish329/kanji-chan/backend/internal/database"
	"github.com/sweetfish329/kanji-chan/backend/internal/model"
)

// HandleListAPIKeys ログイン中のユーザーの API キー一覧を取得
func HandleListAPIKeys(c *echo.Context) error {
	claims, ok := GetUserFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	var apiKeys []model.ApiKey
	if err := database.DB.Where("user_id = ?", claims.UserID).Order("created_at desc").Find(&apiKeys).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch API keys: "+err.Error())
	}

	return c.JSON(http.StatusOK, apiKeys)
}

// HandleCreateAPIKey 新しい API キーを発行
func HandleCreateAPIKey(c *echo.Context) error {
	claims, ok := GetUserFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	var req struct {
		Name string `json:"name"`
	}
	_ = c.Bind(&req)
	if req.Name == "" {
		req.Name = "デフォルト API キー"
	}

	rawKey, apiKey, err := auth.GenerateAPIKey(req.Name, claims.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate API key: "+err.Error())
	}

	// 生成時のみ生の key を含めて返す
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"id":         apiKey.ID,
		"user_id":    apiKey.UserID,
		"name":       apiKey.Name,
		"key":        rawKey, // 一度だけフロントエンドに表示
		"key_prefix": apiKey.KeyPrefix,
		"created_at": apiKey.CreatedAt,
	})
}

// HandleDeleteAPIKey API キーの削除
func HandleDeleteAPIKey(c *echo.Context) error {
	claims, ok := GetUserFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid API key ID")
	}

	result := database.DB.Where("id = ? AND user_id = ?", uint(id), claims.UserID).Delete(&model.ApiKey{})
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete API key: "+result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "API key not found")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "API key deleted successfully"})
}
