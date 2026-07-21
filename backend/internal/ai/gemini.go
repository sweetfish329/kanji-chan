package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sweetfish329/kanji-chan/backend/internal/model"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model/gemini"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
)

const (
	geminiModelName = "gemini-2.5-flash"
)

// ImageInput AIに送信する添付画像データ
type ImageInput struct {
	Data     string `json:"data"`      // Base64文字列 (Data URL "data:image/png;base64,..." または純粋なBase64)
	MimeType string `json:"mime_type"` // e.g. "image/png", "image/jpeg"
}

func parseImageData(img ImageInput) ([]byte, string, error) {
	dataStr := img.Data
	mime := img.MimeType
	if strings.Contains(dataStr, ";base64,") {
		parts := strings.SplitN(dataStr, ";base64,", 2)
		if strings.HasPrefix(parts[0], "data:") {
			mime = strings.TrimPrefix(parts[0], "data:")
		}
		dataStr = parts[1]
	}
	b, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		return nil, "", err
	}
	if mime == "" {
		mime = "image/png"
	}
	return b, mime, nil
}

// getAPIKey 幹事ユーザーが個別に設定したAPIキーを取得 (環境変数へのフォールバックは行わない)
func getAPIKey(user *model.User) string {
	if user != nil {
		return user.GeminiAPIKey
	}
	return ""
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

// ParseEvent 自然文および画像からイベント名と候補日を抽出する
func ParseEvent(ctx context.Context, text string, images []ImageInput, user *model.User) (*ParsedEventResponse, error) {
	apiKey := getAPIKey(user)
	if apiKey == "" {
		return nil, fmt.Errorf("Gemini APIキーが設定されていません。設定画面からGemini APIキーを登録するとAI機能を利用できます。")
	}

	// 1. モデルの初期化
	modelClient, err := gemini.NewModel(ctx, geminiModelName, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini model: %w", err)
	}

	// 2. スキーマ構築
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
	systemPrompt := fmt.Sprintf(`You are a scheduler AI helper. Your task is to analyze the user's request (text prompt and/or attached image files such as event flyers, calendar screenshots, schedule memos, or menu photos) and extract:
1. A concise, nice title for the event.
2. A description of the event (include details found in text or image).
3. Candidate date and time slots based on the text or image content.

Current date is: %s.
Note:
- If the user or image says "next week", calculate the dates based on the current date.
- Default to 2 hours duration if not specified.
- Generate at least 2-4 candidate slots as requested or reasonable based on the flyer/image info.`, currentTime)

	// 3. エージェントの初期化
	eventAgent, err := llmagent.New(llmagent.Config{
		Name:        "event_parser_agent",
		Model:       modelClient,
		Description: "Parses natural language and images into event candidates.",
		Instruction: systemPrompt,
		GenerateContentConfig: &genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema:   eventSchema,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM Agent: %w", err)
	}

	// 4. ランナーの初期化と実行
	sessionService := session.InMemoryService()
	r, err := runner.New(runner.Config{
		Agent:             eventAgent,
		SessionService:    sessionService,
		AutoCreateSession: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create runner: %w", err)
	}

	var parts []*genai.Part
	if text != "" {
		parts = append(parts, genai.NewPartFromText(text))
	}
	for _, img := range images {
		b, mime, err := parseImageData(img)
		if err == nil && len(b) > 0 {
			parts = append(parts, genai.NewPartFromBytes(b, mime))
		}
	}
	if len(parts) == 0 {
		parts = append(parts, genai.NewPartFromText("Extract event info."))
	}

	inputMsg := &genai.Content{
		Role:  genai.RoleUser,
		Parts: parts,
	}

	events := r.Run(ctx, "organizer", "parse-session", inputMsg, agent.RunConfig{})

	var responseText string
	for ev, err := range events {
		if err != nil {
			return nil, fmt.Errorf("runner execution error: %w", err)
		}
		if ev.Content != nil && len(ev.Content.Parts) > 0 {
			for _, part := range ev.Content.Parts {
				if part.Text != "" {
					responseText = part.Text
					break
				}
			}
		}
	}

	if responseText == "" {
		return nil, fmt.Errorf("empty response from ADK agent")
	}

	var parsed ParsedEventResponse
	if err := json.Unmarshal([]byte(responseText), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w (raw: %s)", err, responseText)
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
func SuggestSchedule(ctx context.Context, event *model.Event, preferences string, images []ImageInput, user *model.User) (*AISuggestionResponse, error) {
	apiKey := getAPIKey(user)
	if apiKey == "" {
		return nil, fmt.Errorf("Gemini APIキーが設定されていません。設定画面からGemini APIキーを登録するとAI機能を利用できます。")
	}

	// 1. モデルの初期化
	modelClient, err := gemini.NewModel(ctx, geminiModelName, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini model: %w", err)
	}

	// イベント、候補日、回答データをプロンプト用に整形
	eventDataJSON, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return nil, err
	}

	// 2. スキーマ構築
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

	systemPrompt := `You are an expert scheduler coordinator. Analyze the provided Event data which contains candidates (slots) and user responses, plus any organizer text preferences or attached image notes/memos/calendar screenshots.
Each response has answers mapping to the candidates with status 'ok' (circle), 'maybe' (triangle), or 'ng' (cross).

Your task:
1. Score each candidate slot. (Standard scoring: 'ok' = 2 points, 'maybe' = 1 point, 'ng' = 0 points)
2. Recommend the best 3 slots (or all slots if less than 3).
3. Take into account the organizer's custom preferences and any attached images (e.g. 'A is a key person', 'Prefer weekdays', image notes).
4. Provide a detailed explanation for each rank (who cannot attend, pros/cons).
5. Write the response in Japanese.`

	// 3. エージェントの初期化
	suggestAgent, err := llmagent.New(llmagent.Config{
		Name:        "scheduler_assistant_agent",
		Model:       modelClient,
		Description: "Analyzes schedule responses and coordinates the best slot.",
		Instruction: systemPrompt,
		GenerateContentConfig: &genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema:   suggestionSchema,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM Agent: %w", err)
	}

	// 4. ランナーの初期化と実行
	sessionService := session.InMemoryService()
	r, err := runner.New(runner.Config{
		Agent:             suggestAgent,
		SessionService:    sessionService,
		AutoCreateSession: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create runner: %w", err)
	}

	promptText := fmt.Sprintf("Organizer Preferences: %s\n\nEvent Responses: %s", preferences, string(eventDataJSON))
	var parts []*genai.Part
	parts = append(parts, genai.NewPartFromText(promptText))
	for _, img := range images {
		b, mime, err := parseImageData(img)
		if err == nil && len(b) > 0 {
			parts = append(parts, genai.NewPartFromBytes(b, mime))
		}
	}

	inputMsg := &genai.Content{
		Role:  genai.RoleUser,
		Parts: parts,
	}
	events := r.Run(ctx, "organizer", "suggest-session", inputMsg, agent.RunConfig{})

	var responseText string
	for ev, err := range events {
		if err != nil {
			return nil, fmt.Errorf("runner execution error: %w", err)
		}
		if ev.Content != nil && len(ev.Content.Parts) > 0 {
			for _, part := range ev.Content.Parts {
				if part.Text != "" {
					responseText = part.Text
					break
				}
			}
		}
	}

	if responseText == "" {
		return nil, fmt.Errorf("empty response from ADK agent")
	}

	var suggestions AISuggestionResponse
	if err := json.Unmarshal([]byte(responseText), &suggestions); err != nil {
		return nil, fmt.Errorf("failed to parse AI suggestions: %w (raw: %s)", err, responseText)
	}

	return &suggestions, nil
}
