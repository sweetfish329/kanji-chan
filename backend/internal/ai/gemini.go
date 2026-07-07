package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sweetfish329/kanji-chan/backend/internal/model"
)

const (
	geminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"
)

// Gemini API Request structs
type Part struct {
	Text string `json:"text"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Schema struct {
	Type        string            `json:"type"`
	Properties  map[string]Schema `json:"properties,omitempty"`
	Required    []string          `json:"required,omitempty"`
	Items       *Schema           `json:"items,omitempty"`
	Description string            `json:"description,omitempty"`
}

type ResponseMimeType string

const (
	MimeTypeJSON ResponseMimeType = "application/json"
)

type Configuration struct {
	ResponseMimeType ResponseMimeType `json:"responseMimeType,omitempty"`
	ResponseSchema   *Schema          `json:"responseSchema,omitempty"`
}

type GeminiRequest struct {
	Contents         []Content      `json:"contents"`
	GenerationConfig Configuration  `json:"generationConfig,omitempty"`
}

// Gemini API Response structs
type Candidate struct {
	Content struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"content"`
}

type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

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

	// Schema for structured JSON output
	eventSchema := &Schema{
		Type: "OBJECT",
		Properties: map[string]Schema{
			"title": {Type: "STRING", Description: "The summarized event name (e.g. 'Shibuya Drinking Party')"},
			"description": {Type: "STRING", Description: "The extracted purpose or summary of the event"},
			"candidates": {
				Type: "ARRAY",
				Items: &Schema{
					Type: "OBJECT",
					Properties: map[string]Schema{
						"event_date": {Type: "STRING", Description: "Suggested date in YYYY-MM-DD format"},
						"start_time": {Type: "STRING", Description: "Suggested start time in HH:MM format (24-hour)"},
						"end_time":   {Type: "STRING", Description: "Suggested end time in HH:MM format (24-hour). If not specified, default to 2 hours after start_time"},
					},
					Required: []string{"event_date", "start_time", "end_time"},
				},
				Description: "List of proposed candidate slots",
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

	reqPayload := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: fmt.Sprintf("%s\n\nUser request: %s", systemPrompt, text)},
				},
			},
		},
		GenerationConfig: Configuration{
			ResponseMimeType: MimeTypeJSON,
			ResponseSchema:   eventSchema,
		},
	}

	respBody, err := callGeminiAPI(ctx, apiKey, reqPayload)
	if err != nil {
		return nil, err
	}

	var parsed ParsedEventResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w (raw: %s)", err, string(respBody))
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

	// イベント、候補日、回答データをプロンプト用に整形
	eventDataJSON, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return nil, err
	}

	// Schema for structured JSON output
	suggestionSchema := &Schema{
		Type: "OBJECT",
		Properties: map[string]Schema{
			"suggestions": {
				Type: "ARRAY",
				Items: &Schema{
					Type: "OBJECT",
					Properties: map[string]Schema{
						"candidate_id": {Type: "INTEGER", Description: "The ID of the candidate slot"},
						"rank":         {Type: "INTEGER", Description: "Recommendation rank (1=Best, 2=Second, 3=Third)"},
						"score":        {Type: "INTEGER", Description: "Calculated compatibility score (e.g., ok=2pts, maybe=1pt, ng=0pts)"},
						"reason":       {Type: "STRING", Description: "Detailed reason for the recommendation (e.g., 'All 5 people can join', 'Maximum participation but Person X cannot attend')"},
					},
					Required: []string{"candidate_id", "rank", "score", "reason"},
				},
			},
			"overall_analysis": {Type: "STRING", Description: "Overall summary analysis and strategic advice for the organizer"},
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

	reqPayload := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: fmt.Sprintf("%s\n\nOrganizer Preferences: %s\n\nEvent Responses: %s", systemPrompt, preferences, string(eventDataJSON))},
				},
			},
		},
		GenerationConfig: Configuration{
			ResponseMimeType: MimeTypeJSON,
			ResponseSchema:   suggestionSchema,
		},
	}

	respBody, err := callGeminiAPI(ctx, apiKey, reqPayload)
	if err != nil {
		return nil, err
	}

	var suggestions AISuggestionResponse
	if err := json.Unmarshal(respBody, &suggestions); err != nil {
		return nil, fmt.Errorf("failed to parse AI suggestions: %w (raw: %s)", err, string(respBody))
	}

	return &suggestions, nil
}

// callGeminiAPI Gemini APIとの実際の通信部分
func callGeminiAPI(ctx context.Context, apiKey string, payload GeminiRequest) ([]byte, error) {
	url := fmt.Sprintf("%s?key=%s", geminiAPIURL, apiKey)
	reqJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errData map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errData)
		return nil, fmt.Errorf("API returned non-200 status %d: %v", resp.StatusCode, errData)
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini API")
	}

	return []byte(geminiResp.Candidates[0].Content.Parts[0].Text), nil
}
