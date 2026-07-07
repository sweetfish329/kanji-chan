package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/sweetfish329/kanji-chan/backend/internal/ai"
	"github.com/sweetfish329/kanji-chan/backend/internal/database"
	"github.com/sweetfish329/kanji-chan/backend/internal/model"
)

// HandleParseEvent 自然文からイベント候補日を解析 (幹事専用)
func HandleParseEvent(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetUserFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Text == "" {
		writeError(w, http.StatusBadRequest, "Text input is required")
		return
	}

	// 幹事ユーザー情報を取得
	var user model.User
	if err := database.DB.First(&user, claims.UserID).Error; err != nil {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	parsed, err := ai.ParseEvent(r.Context(), req.Text, &user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "AI parsing failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, parsed)
}

// HandleSuggestSchedule 回答状況から最適な日程を絞り込む (幹事専用)
func HandleSuggestSchedule(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetUserFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		EventID     string `json:"event_id"`
		Preferences string `json:"preferences"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	eventID, err := uuid.Parse(req.EventID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid event UUID format")
		return
	}

	// イベントデータの取得 (Candidates, Responses と Answers をすべてロード)
	var event model.Event
	err = database.DB.
		Preload("Candidates").
		Preload("Responses").
		Preload("Responses.Answers").
		First(&event, "id = ?", eventID).Error

	if err != nil {
		writeError(w, http.StatusNotFound, "Event not found")
		return
	}

	// 権限チェック (作成者のみ)
	if event.CreatedBy == nil || *event.CreatedBy != claims.UserID {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}

	// 幹事ユーザー情報を取得
	var user model.User
	if err := database.DB.First(&user, claims.UserID).Error; err != nil {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	suggestions, err := ai.SuggestSchedule(r.Context(), &event, req.Preferences, &user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "AI suggestion failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, suggestions)
}
