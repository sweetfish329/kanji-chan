package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/sweetfish329/kanji-chan/backend/internal/model"
	"google.golang.org/api/option"
)

const (
	geminiModelName = "gemini-2.5-flash"
)

// getAPIKey 優先順位: 幹事ユーザーが個別に設定したAPIキー > 環境変数
func getAPIKey(user *model.User) string {
	if user != nil && user.GeminiAPIKey != "" {
		return user.GeminiAPIKey
	}
	return os.Getenv("GEMINI_API_KEY")
}

// ParsedEventResponse 自然文パース結果のフロントエンド用レスポンス
type ParsedEventResponse struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Candidates  []struct {
		EventDate string `json:"event_date"` // YYYY-MM-DD
		StartTime string `json:"start_time"` // HH:MM
		EndTime   string `json:"end_time"`   // HH:MM
	} `json:"candidates"`
}

// ParseEvent 自然文からイベント名と候補日を抽出する
func ParseEvent(ctx context.Context, text string, user *model.User) (*ParsedEventResponse, error) {
	apiKey := getAPIKey(user)
	if apiKey == "" {
		return nil, fmt.Errorf("Gemini API key is not configured")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(geminiModelName)

	// Go SDK の Schema 構築
	eventSchema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"title":       {Type: genai.TypeString},
			"description": {Type: genai.TypeString},
			"candidates": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"event_date": {Type: genai.TypeString},
						"start_time": {Type: genai.TypeString},
						"end_time":   {Type: genai.TypeString},
					},
					Required: []string{"event_date", "start_time", "end_time"},
				},
			},
		},
		Required: []string{"title", "description", "candidates"},
	}

	currentTime := time.Now().Format("2006-01-02 (Monday)")
	systemPrompt := fmt.Sprintf(`You are a scheduler AI helper. Your task is to analyze the user's natural language request and extract:
1. A concise, nice title for the event.
2. A description of the event.
3. Candidate date and time slots based on the text.

Current date is: %s.
Note:
- If the user says "next week", calculate the dates based on the current date.
- Default to 2 hours duration if not specified.
- Generate at least 2-4 candidate slots as requested or reasonable.`, currentTime)

	model.SystemInstruction = genai.NewUserContent(genai.Text(systemPrompt))
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = eventSchema

	resp, err := model.GenerateContent(ctx, genai.Text(text))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini API")
	}

	respPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return nil, fmt.Errorf("unexpected response part type")
	}

	var parsed ParsedEventResponse
	if err := json.Unmarshal([]byte(respPart), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w (raw: %s)", err, string(respPart))
	}

	return &parsed, nil
}

// AIRecommendation 提案候補日
type AIRecommendation struct {
	CandidateID uint   `json:"candidate_id"`
	Rank        int    `json:"rank"`
	Score       int    `json:"score"`
	Reason      string `json:"reason"`
}

// AISuggestionResponse AIの予定調整提案
type AISuggestionResponse struct {
	Suggestions     []AIRecommendation `json:"suggestions"`
	OverallAnalysis string             `json:"overall_analysis"`
}

// SuggestSchedule 回答結果から候補日を絞り込む
func SuggestSchedule(ctx context.Context, event *model.Event, preferences string, user *model.User) (*AISuggestionResponse, error) {
	apiKey := getAPIKey(user)
	if apiKey == "" {
		return nil, fmt.Errorf("Gemini API key is not configured")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(geminiModelName)

	// イベント、候補日、回答データをプロンプト用に整形
	eventDataJSON, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return nil, err
	}

	// Go SDK の Schema 構築
	suggestionSchema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"suggestions": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"candidate_id": {Type: genai.TypeInteger},
						"rank":         {Type: genai.TypeInteger},
						"score":        {Type: genai.TypeInteger},
						"reason":       {Type: genai.TypeString},
					},
					Required: []string{"candidate_id", "rank", "score", "reason"},
				},
			},
			"overall_analysis": {Type: genai.TypeString},
		},
		Required: []string{"suggestions", "overall_analysis"},
	}

	systemPrompt := `You are an expert scheduler coordinator. Analyze the provided Event data which contains candidates (slots) and user responses.
Each response has answers mapping to the candidates with status 'ok' (circle), 'maybe' (triangle), or 'ng' (cross).

Your task:
1. Score each candidate slot. (Standard scoring: 'ok' = 2 points, 'maybe' = 1 point, 'ng' = 0 points)
2. Recommend the best 3 slots (or all slots if less than 3).
3. Take into account the organizer's custom preferences (e.g. 'A is a key person', 'Prefer weekdays').
4. Provide a detailed explanation for each rank (who cannot attend, pros/cons).
5. Write the response in Japanese.`

	model.SystemInstruction = genai.NewUserContent(genai.Text(systemPrompt))
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = suggestionSchema

	promptText := fmt.Sprintf("Organizer Preferences: %s\n\nEvent Responses: %s", preferences, string(eventDataJSON))
	resp, err := model.GenerateContent(ctx, genai.Text(promptText))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini API")
	}

	respPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return nil, fmt.Errorf("unexpected response part type")
	}

	var suggestions AISuggestionResponse
	if err := json.Unmarshal([]byte(respPart), &suggestions); err != nil {
		return nil, fmt.Errorf("failed to parse AI suggestions: %w (raw: %s)", err, string(respPart))
	}

	return &suggestions, nil
}
