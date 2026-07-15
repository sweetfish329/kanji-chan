package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/sweetfish329/kanji-chan/backend/internal/database"
	"github.com/sweetfish329/kanji-chan/backend/internal/model"
)

// Request schemas
type CreateEventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Candidates  []struct {
		EventDate string `json:"event_date"` // YYYY-MM-DD
		StartTime string `json:"start_time"` // HH:MM
		EndTime   string `json:"end_time"`   // HH:MM
	} `json:"candidates"`
}

// HandleCreateEvent 新規イベント作成 (ログイン不要・匿名作成も可)
func HandleCreateEvent(c *echo.Context) error {
	var createdByID *uint
	claims, ok := GetUserFromContext(c)
	if ok {
		createdByID = &claims.UserID
	}

	var req CreateEventRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.Title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Title is required")
	}

	// トランザクションで作成
	tx := database.DB.Begin()

	eventID := uuid.New()
	event := model.Event{
		ID:          eventID,
		Title:       req.Title,
		Description: req.Description,
		CreatedBy:   createdByID,
		Status:      "scheduling",
	}

	if err := tx.Create(&event).Error; err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create event: "+err.Error())
	}

	for _, cand := range req.Candidates {
		candidate := model.EventCandidate{
			EventID:   eventID,
			EventDate: cand.EventDate,
			StartTime: cand.StartTime,
			EndTime:   cand.EndTime,
		}
		if err := tx.Create(&candidate).Error; err != nil {
			tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create event candidates: "+err.Error())
		}
	}

	if err := tx.Commit().Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}

	// 作成したイベントをリレーション込みで再取得して返す
	var createdEvent model.Event
	database.DB.Preload("Candidates").First(&createdEvent, "id = ?", eventID)
	return c.JSON(http.StatusCreated, createdEvent)
}

// HandleListEvents ログイン中の幹事が作成したイベント一覧を取得
func HandleListEvents(c *echo.Context) error {
	claims, ok := GetUserFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	var events []model.Event
	err := database.DB.Where("created_by = ?", claims.UserID).Order("created_at desc").Find(&events).Error
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch events: "+err.Error())
	}

	return c.JSON(http.StatusOK, events)
}

// HandleGetEvent イベント詳細を取得 (回答状況・候補日含む、ログイン不要)
func HandleGetEvent(c *echo.Context) error {
	eventIDStr := c.Param("id")
	if eventIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing event ID")
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format")
	}

	var event model.Event
	// Candidates と Responses (とその中の Answers) を一括プリロード
	err = database.DB.
		Preload("Candidates").
		Preload("Responses").
		Preload("Responses.Answers").
		Preload("ConfirmedCandidate").
		First(&event, "id = ?", eventID).Error

	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Event not found")
	}

	return c.JSON(http.StatusOK, event)
}

// HandleUpdateEvent イベント情報の更新・確定 (幹事専用)
func HandleUpdateEvent(c *echo.Context) error {
	claims, ok := GetUserFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	eventIDStr := c.Param("id")
	if eventIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing event ID")
	}
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format")
	}

	var event model.Event
	if err := database.DB.First(&event, "id = ?", eventID).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Event not found")
	}

	// 権限チェック (作成者のみ)
	if event.CreatedBy == nil || *event.CreatedBy != claims.UserID {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	var req struct {
		Title                string `json:"title"`
		Description          string `json:"description"`
		Status               string `json:"status"` // 'scheduling' or 'confirmed'
		ConfirmedCandidateID *uint  `json:"confirmed_candidate_id"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.Title != "" {
		event.Title = req.Title
	}
	event.Description = req.Description
	if req.Status != "" {
		event.Status = req.Status
	}
	event.ConfirmedCandidateID = req.ConfirmedCandidateID

	if err := database.DB.Save(&event).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update event: "+err.Error())
	}

	// 更新後の状態を取得
	database.DB.Preload("Candidates").Preload("ConfirmedCandidate").First(&event, "id = ?", eventID)
	return c.JSON(http.StatusOK, event)
}

// HandleDeleteEvent イベントの削除 (幹事専用)
func HandleDeleteEvent(c *echo.Context) error {
	claims, ok := GetUserFromContext(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	eventIDStr := c.Param("id")
	if eventIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing event ID")
	}
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format")
	}

	var event model.Event
	if err := database.DB.First(&event, "id = ?", eventID).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Event not found")
	}

	if event.CreatedBy == nil || *event.CreatedBy != claims.UserID {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	if err := database.DB.Delete(&event).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete event: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Event deleted successfully"})
}
