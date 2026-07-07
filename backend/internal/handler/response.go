package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/sweetfish329/kanji-chan/backend/internal/database"
	"github.com/sweetfish329/kanji-chan/backend/internal/model"
)

// Request schemas
type AnswerRequest struct {
	CandidateID  uint   `json:"candidate_id"`
	AnswerStatus string `json:"answer_status"` // 'ok', 'maybe', 'ng'
}

type AddResponseRequest struct {
	RespondentName string          `json:"respondent_name"`
	Comment        string          `json:"comment"`
	Answers        []AnswerRequest `json:"answers"`
}

// HandleAddResponse イベントに対する回答の登録 (ログイン不要)
func HandleAddResponse(w http.ResponseWriter, r *http.Request) {
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

	// イベント存在チェック
	var event model.Event
	if err := database.DB.First(&event, "id = ?", eventID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Event not found")
		return
	}

	var req AddResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RespondentName == "" {
		writeError(w, http.StatusBadRequest, "Respondent name is required")
		return
	}

	// トランザクション処理
	tx := database.DB.Begin()

	response := model.Response{
		EventID:        eventID,
		RespondentName: req.RespondentName,
		Comment:        req.Comment,
	}

	if err := tx.Create(&response).Error; err != nil {
		tx.Rollback()
		writeError(w, http.StatusInternalServerError, "Failed to create response: "+err.Error())
		return
	}

	for _, ans := range req.Answers {
		answer := model.CandidateAnswer{
			ResponseID:   response.ID,
			CandidateID:  ans.CandidateID,
			AnswerStatus: ans.AnswerStatus,
		}
		if err := tx.Create(&answer).Error; err != nil {
			tx.Rollback()
			writeError(w, http.StatusInternalServerError, "Failed to create candidate answer: "+err.Error())
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	// 再取得して返す
	var createdResponse model.Response
	database.DB.Preload("Answers").First(&createdResponse, response.ID)
	writeJSON(w, http.StatusCreated, createdResponse)
}

// HandleDeleteResponse 回答の削除 (調整さんライクにID指定で誰でも、または幹事のみ。ここではシンプルにID指定で削除)
func HandleDeleteResponse(w http.ResponseWriter, r *http.Request) {
	responseIDStr := r.PathValue("response_id")
	if responseIDStr == "" {
		writeError(w, http.StatusBadRequest, "Missing response ID")
		return
	}
	responseID, err := strconv.Atoi(responseIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid response ID format")
		return
	}

	var response model.Response
	if err := database.DB.First(&response, responseID).Error; err != nil {
		writeError(w, http.StatusNotFound, "Response not found")
		return
	}

	// ※ 幹事認証を入れる場合は、claimsを取得し event.CreatedBy と比較するが、
	// 調整さんのような緩い削除を許容するため、まずはシンプルに削除実行
	if err := database.DB.Delete(&response).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete response: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Response deleted successfully"})
}
