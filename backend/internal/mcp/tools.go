package mcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sweetfish329/kanji-chan/backend/internal/database"
	"github.com/sweetfish329/kanji-chan/backend/internal/model"
)

// ListEventsInput
type ListEventsInput struct{}

type EventSummary struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

type ListEventsOutput struct {
	Events []EventSummary `json:"events"`
}

// GetEventInput
type GetEventInput struct {
	EventID string `json:"event_id" jsonschema:"対象イベントのUUID"`
}

type GetEventOutput struct {
	Event model.Event `json:"event"`
}

// CreateEventInput
type CandidateInput struct {
	EventDate string `json:"event_date" jsonschema:"開催日 (YYYY-MM-DD形式)"`
	StartTime string `json:"start_time" jsonschema:"開始時間 (HH:MM形式)"`
	EndTime   string `json:"end_time" jsonschema:"終了時間 (HH:MM形式)"`
}

type CreateEventInput struct {
	Title       string           `json:"title" jsonschema:"イベントタイトル"`
	Description string           `json:"description,omitempty" jsonschema:"イベントの詳細説明"`
	Candidates  []CandidateInput `json:"candidates" jsonschema:"候補日時のリスト (1つ以上必要)"`
}

type CreateEventOutput struct {
	Event model.Event `json:"event"`
	Msg   string      `json:"message"`
}

// UpdateEventInput
type UpdateEventInput struct {
	EventID              string `json:"event_id" jsonschema:"対象イベントのUUID"`
	Title                string `json:"title,omitempty" jsonschema:"新しいイベントタイトル"`
	Description          string `json:"description,omitempty" jsonschema:"新しい詳細説明"`
	Status               string `json:"status,omitempty" jsonschema:"ステータス ('scheduling' または 'confirmed')"`
	ConfirmedCandidateID *uint  `json:"confirmed_candidate_id,omitempty" jsonschema:"最終確定候補日時のID"`
}

type UpdateEventOutput struct {
	Event model.Event `json:"event"`
	Msg   string      `json:"message"`
}

// DeleteEventInput
type DeleteEventInput struct {
	EventID string `json:"event_id" jsonschema:"削除するイベントのUUID"`
}

type DeleteEventOutput struct {
	Success bool   `json:"success"`
	Msg     string `json:"message"`
}

// RegisterTools MCPサーバーにすべてのツールを登録
func RegisterTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_events",
		Description: "ログインユーザーが作成した日程調整イベントの一覧を取得します。",
	}, handleListEvents)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_event",
		Description: "指定されたイベントIDの詳細（候補日時、回答者一覧・回答状況、確定日時など）を取得します。",
	}, handleGetEvent)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_event",
		Description: "新しい日程調整イベントを作成します。タイトル、説明、候補日時スロットを指定できます。",
	}, handleCreateEvent)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_event",
		Description: "既存の日程調整イベントのタイトル、説明の変更や開催候補日時の確定を行ないます。",
	}, handleUpdateEvent)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_event",
		Description: "指定されたイベントを削除します。",
	}, handleDeleteEvent)
}

func handleListEvents(ctx context.Context, req *mcp.CallToolRequest, in ListEventsInput) (*mcp.CallToolResult, ListEventsOutput, error) {
	claims, ok := GetClaimsFromContext(ctx)
	if !ok {
		return nil, ListEventsOutput{}, errors.New("authentication required: valid API key or session token needed")
	}

	var events []model.Event
	if err := database.DB.Where("created_by = ?", claims.UserID).Order("created_at desc").Find(&events).Error; err != nil {
		return nil, ListEventsOutput{}, fmt.Errorf("failed to fetch events: %w", err)
	}

	summaries := make([]EventSummary, 0, len(events))
	for _, e := range events {
		summaries = append(summaries, EventSummary{
			ID:          e.ID,
			Title:       e.Title,
			Description: e.Description,
			Status:      e.Status,
			CreatedAt:   e.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return nil, ListEventsOutput{Events: summaries}, nil
}

func handleGetEvent(ctx context.Context, req *mcp.CallToolRequest, in GetEventInput) (*mcp.CallToolResult, GetEventOutput, error) {
	claims, ok := GetClaimsFromContext(ctx)
	if !ok {
		return nil, GetEventOutput{}, errors.New("authentication required: valid API key or session token needed")
	}

	if in.EventID == "" {
		return nil, GetEventOutput{}, errors.New("event_id is required")
	}

	var event model.Event
	err := database.DB.Preload("Candidates").
		Preload("Responses.Answers").
		Preload("ConfirmedCandidate").
		Where("id = ? AND created_by = ?", in.EventID, claims.UserID).
		First(&event).Error

	if err != nil {
		return nil, GetEventOutput{}, fmt.Errorf("event not found or access denied for ID: %s", in.EventID)
	}

	return nil, GetEventOutput{Event: event}, nil
}

func handleCreateEvent(ctx context.Context, req *mcp.CallToolRequest, in CreateEventInput) (*mcp.CallToolResult, CreateEventOutput, error) {
	claims, ok := GetClaimsFromContext(ctx)
	if !ok {
		return nil, CreateEventOutput{}, errors.New("authentication required: valid API key or session token needed")
	}

	if in.Title == "" {
		return nil, CreateEventOutput{}, errors.New("title is required")
	}
	if len(in.Candidates) == 0 {
		return nil, CreateEventOutput{}, errors.New("at least one candidate date slot is required")
	}

	event := model.Event{
		Title:       in.Title,
		Description: in.Description,
		CreatedBy:   &claims.UserID,
		Status:      "scheduling",
	}

	for _, c := range in.Candidates {
		if c.EventDate == "" {
			return nil, CreateEventOutput{}, errors.New("event_date is required for each candidate")
		}
		startTime := c.StartTime
		if startTime == "" {
			startTime = "19:00"
		}
		endTime := c.EndTime
		if endTime == "" {
			endTime = "21:00"
		}
		event.Candidates = append(event.Candidates, model.EventCandidate{
			EventDate: c.EventDate,
			StartTime: startTime,
			EndTime:   endTime,
		})
	}

	if err := database.DB.Create(&event).Error; err != nil {
		return nil, CreateEventOutput{}, fmt.Errorf("failed to create event: %w", err)
	}

	return nil, CreateEventOutput{
		Event: event,
		Msg:   fmt.Sprintf("Successfully created event '%s' with ID %s", event.Title, event.ID),
	}, nil
}

func handleUpdateEvent(ctx context.Context, req *mcp.CallToolRequest, in UpdateEventInput) (*mcp.CallToolResult, UpdateEventOutput, error) {
	claims, ok := GetClaimsFromContext(ctx)
	if !ok {
		return nil, UpdateEventOutput{}, errors.New("authentication required: valid API key or session token needed")
	}

	if in.EventID == "" {
		return nil, UpdateEventOutput{}, errors.New("event_id is required")
	}

	var event model.Event
	if err := database.DB.Where("id = ? AND created_by = ?", in.EventID, claims.UserID).First(&event).Error; err != nil {
		return nil, UpdateEventOutput{}, fmt.Errorf("event not found or access denied: %s", in.EventID)
	}

	if in.Title != "" {
		event.Title = in.Title
	}
	if in.Description != "" {
		event.Description = in.Description
	}
	if in.Status != "" {
		event.Status = in.Status
	}
	if in.ConfirmedCandidateID != nil {
		event.ConfirmedCandidateID = in.ConfirmedCandidateID
		event.Status = "confirmed"
	}

	if err := database.DB.Save(&event).Error; err != nil {
		return nil, UpdateEventOutput{}, fmt.Errorf("failed to update event: %w", err)
	}

	// 更新後の情報を候補日つきで再取得
	database.DB.Preload("Candidates").Preload("ConfirmedCandidate").First(&event, "id = ?", event.ID)

	return nil, UpdateEventOutput{
		Event: event,
		Msg:   fmt.Sprintf("Successfully updated event '%s' (%s)", event.Title, event.ID),
	}, nil
}

func handleDeleteEvent(ctx context.Context, req *mcp.CallToolRequest, in DeleteEventInput) (*mcp.CallToolResult, DeleteEventOutput, error) {
	claims, ok := GetClaimsFromContext(ctx)
	if !ok {
		return nil, DeleteEventOutput{}, errors.New("authentication required: valid API key or session token needed")
	}

	if in.EventID == "" {
		return nil, DeleteEventOutput{}, errors.New("event_id is required")
	}

	result := database.DB.Where("id = ? AND created_by = ?", in.EventID, claims.UserID).Delete(&model.Event{})
	if result.Error != nil {
		return nil, DeleteEventOutput{}, fmt.Errorf("failed to delete event: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, DeleteEventOutput{}, fmt.Errorf("event not found or access denied: %s", in.EventID)
	}

	return nil, DeleteEventOutput{
		Success: true,
		Msg:     fmt.Sprintf("Successfully deleted event ID %s", in.EventID),
	}, nil
}
