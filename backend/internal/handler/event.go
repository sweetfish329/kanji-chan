package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
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

// HandleCreateEvent 新規イベント作成 (幹事専用)
func HandleCreateEvent(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetUserFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "Title is required")
		return
	}

	// トランザクションで作成
	tx := database.DB.Begin()

	eventID := uuid.New()
	event := model.Event{
		ID:          eventID,
		Title:       req.Title,
		Description: req.Description,
		CreatedBy:   &claims.UserID,
		Status:      "scheduling",
	}

	if err := tx.Create(&event).Error; err != nil {
		tx.Rollback()
		writeError(w, http.StatusInternalServerError, "Failed to create event: "+err.Error())
		return
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
			writeError(w, http.StatusInternalServerError, "Failed to create event candidates: "+err.Error())
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	// 作成したイベントをリレーション込みで再取得して返す
	var createdEvent model.Event
	database.DB.Preload("Candidates").First(&createdEvent, "id = ?", eventID)
	writeJSON(w, http.StatusCreated, createdEvent)
}

// HandleListEvents ログイン中の幹事が作成したイベント一覧を取得
func HandleListEvents(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetUserFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var events []model.Event
	err := database.DB.Where("created_by = ?", claims.UserID).Order("created_at desc").Find(&events).Error
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch events: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, events)
}

// HandleGetEvent イベント詳細を取得 (回答状況・候補日含む、ログイン不要)
func HandleGetEvent(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.PathValue("id")
	if eventIDStr == "" {
		writeError(w, http.StatusBadRequest, "Missing event ID")
		return
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid UUID format")
		return
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
		writeError(w, http.StatusNotFound, "Event not found")
		return
	}

	writeJSON(w, http.StatusOK, event)
}

// HandleUpdateEvent イベント情報の更新・確定 (幹事専用)
func HandleUpdateEvent(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetUserFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	eventIDStr := r.PathValue("id")
	if eventIDStr == "" {
		writeError(w, http.StatusBadRequest, "Missing event ID")
		return
	}
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid UUID format")
		return
	}

	var event model.Event
	if err := database.DB.First(&event, "id = ?", eventID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Event not found")
		return
	}

	// 権限チェック (作成者のみ)
	if event.CreatedBy == nil || *event.CreatedBy != claims.UserID {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}

	var req struct {
		Title                string `json:"title"`
		Description          string `json:"description"`
		Status               string `json:"status"` // 'scheduling' or 'confirmed'
		ConfirmedCandidateID *uint  `json:"confirmed_candidate_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
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
		writeError(w, http.StatusInternalServerError, "Failed to update event: "+err.Error())
		return
	}

	// 更新後の状態を取得
	database.DB.Preload("Candidates").Preload("ConfirmedCandidate").First(&event, "id = ?", eventID)
	writeJSON(w, http.StatusOK, event)
}

// HandleDeleteEvent イベントの削除 (幹事専用)
func HandleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetUserFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	eventIDStr := r.PathValue("id")
	if eventIDStr == "" {
		writeError(w, http.StatusBadRequest, "Missing event ID")
		return
	}
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid UUID format")
		return
	}

	var event model.Event
	if err := database.DB.First(&event, "id = ?", eventID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Event not found")
		return
	}

	if event.CreatedBy == nil || *event.CreatedBy != claims.UserID {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}

	if err := database.DB.Delete(&event).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete event: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Event deleted successfully"})
}
