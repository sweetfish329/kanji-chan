package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/sweetfish329/kanji-chan/backend/internal/ai"
	"github.com/sweetfish329/kanji-chan/backend/internal/database"
	"github.com/sweetfish329/kanji-chan/backend/internal/model"
)

// HandleParseEvent 自然文および画像からイベント候補日を解析 (幹事専用)
func HandleParseEvent(c *echo.Context) error {
	claims, ok := GetUserFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	var req struct {
		Text   string          `json:"text"`
		Images []ai.ImageInput `json:"images"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.Text == "" && len(req.Images) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Text input or image attachments are required")
	}

	// 幹事ユーザー情報を取得
	var user model.User
	if err := database.DB.First(&user, claims.UserID).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	parsed, err := ai.ParseEvent(c.Request().Context(), req.Text, req.Images, &user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "AI parsing failed: "+err.Error())
	}

	return c.JSON(http.StatusOK, parsed)
}

// HandleSuggestSchedule 回答状況から最適な日程を絞り込む (幹事専用)
func HandleSuggestSchedule(c *echo.Context) error {
	claims, ok := GetUserFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	var req struct {
		EventID     string          `json:"event_id"`
		Preferences string          `json:"preferences"`
		Images      []ai.ImageInput `json:"images"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if _, err := uuid.Parse(req.EventID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid event UUID format")
	}

	// イベントデータの取得 (Candidates, Responses と Answers をすべてロード)
	var event model.Event
	err := database.DB.
		Preload("Candidates").
		Preload("Responses").
		Preload("Responses.Answers").
		First(&event, "id = ?", req.EventID).Error

	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Event not found")
	}

	// 権限チェック (作成者のみ)
	if event.CreatedBy == nil || *event.CreatedBy != claims.UserID {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	// 幹事ユーザー情報を取得
	var user model.User
	if err := database.DB.First(&user, claims.UserID).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	suggestions, err := ai.SuggestSchedule(c.Request().Context(), &event, req.Preferences, req.Images, &user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "AI suggestion failed: "+err.Error())
	}

	return c.JSON(http.StatusOK, suggestions)
}
