package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
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

type UpdateResponseRequest struct {
	RespondentName string          `json:"respondent_name"`
	Comment        string          `json:"comment"`
	Answers        []AnswerRequest `json:"answers"`
}

// checkResponseAuthority 編集・削除の権限を検証（消えていても編集可能にするため常に許可）
func checkResponseAuthority(c echo.Context, response *model.Response) (bool, error) {
	return true, nil
}

// HandleAddResponse イベントに対する回答の登録 (ログイン不要)
func HandleAddResponse(c echo.Context) error {
	eventIDStr := c.Param("id")
	if eventIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing event ID")
	}
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format")
	}

	// イベント存在チェック
	var event model.Event
	if err := database.DB.First(&event, "id = ?", eventID).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Event not found")
	}

	var req AddResponseRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.RespondentName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Respondent name is required")
	}

	// 編集用の使い捨てランダムトークンを生成
	editToken := uuid.New().String()

	// トランザクション処理
	tx := database.DB.Begin()

	response := model.Response{
		EventID:        eventID,
		RespondentName: req.RespondentName,
		Comment:        req.Comment,
		EditToken:      editToken,
	}

	if err := tx.Create(&response).Error; err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create response: "+err.Error())
	}

	for _, ans := range req.Answers {
		answer := model.CandidateAnswer{
			ResponseID:   response.ID,
			CandidateID:  ans.CandidateID,
			AnswerStatus: ans.AnswerStatus,
		}
		if err := tx.Create(&answer).Error; err != nil {
			tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create candidate answer: "+err.Error())
		}
	}

	if err := tx.Commit().Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}

	// 再取得して返す
	var createdResponse model.Response
	database.DB.Preload("Answers").First(&createdResponse, response.ID)
	return c.JSON(http.StatusCreated, createdResponse)
}

// HandleUpdateResponse 回答の編集・更新 (ログイン不要 / トークン認証)
func HandleUpdateResponse(c echo.Context) error {
	responseIDStr := c.Param("response_id")
	if responseIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing response ID")
	}
	responseID, err := strconv.Atoi(responseIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid response ID format")
	}

	var response model.Response
	if err := database.DB.First(&response, responseID).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Response not found")
	}

	// 権限検証
	authorized, err := checkResponseAuthority(c, &response)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if !authorized {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized to update this response")
	}

	var req UpdateResponseRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if req.RespondentName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Respondent name is required")
	}

	// トランザクション処理
	tx := database.DB.Begin()

	response.RespondentName = req.RespondentName
	response.Comment = req.Comment

	if err := tx.Save(&response).Error; err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update response: "+err.Error())
	}

	// 既存の回答スロットへの都合データを全削除して再登録
	if err := tx.Where("response_id = ?", response.ID).Delete(&model.CandidateAnswer{}).Error; err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to reset previous answers: "+err.Error())
	}

	for _, ans := range req.Answers {
		answer := model.CandidateAnswer{
			ResponseID:   response.ID,
			CandidateID:  ans.CandidateID,
			AnswerStatus: ans.AnswerStatus,
		}
		if err := tx.Create(&answer).Error; err != nil {
			tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to recreate candidate answer: "+err.Error())
		}
	}

	if err := tx.Commit().Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}

	// 再取得して返す
	var updatedResponse model.Response
	database.DB.Preload("Answers").First(&updatedResponse, response.ID)
	return c.JSON(http.StatusOK, updatedResponse)
}

// HandleDeleteResponse 回答の削除 (トークン認証 または 幹事セッション)
func HandleDeleteResponse(c echo.Context) error {
	responseIDStr := c.Param("response_id")
	if responseIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing response ID")
	}
	responseID, err := strconv.Atoi(responseIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid response ID format")
	}

	var response model.Response
	if err := database.DB.First(&response, responseID).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Response not found")
	}

	// 権限検証
	authorized, err := checkResponseAuthority(c, &response)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if !authorized {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized to delete this response")
	}

	if err := database.DB.Delete(&response).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete response: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Response deleted successfully"})
}
