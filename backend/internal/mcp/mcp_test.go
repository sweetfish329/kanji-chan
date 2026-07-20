package mcp

import (
	"context"
	"os"
	"testing"

	"github.com/sweetfish329/kanji-chan/backend/internal/auth"
	"github.com/sweetfish329/kanji-chan/backend/internal/database"
	"github.com/sweetfish329/kanji-chan/backend/internal/model"
)

func TestMCPTools(t *testing.T) {
	os.Setenv("DB_PATH", ":memory:")
	_, err := database.InitDB()
	if err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}

	user := model.User{
		OAuthProvider: "google",
		OAuthID:       "test-user-1",
		Email:         "test@example.com",
		Name:          "Test User",
	}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	ctx := context.WithValue(context.Background(), UserClaimsKey, &auth.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
	})

	// 1. Create Event
	createInput := CreateEventInput{
		Title:       "MCPテスト飲み会",
		Description: "MCPから作成されたテストイベントです",
		Candidates: []CandidateInput{
			{EventDate: "2026-08-01", StartTime: "19:00", EndTime: "21:00"},
			{EventDate: "2026-08-02", StartTime: "19:00", EndTime: "21:00"},
		},
	}

	_, createRes, err := handleCreateEvent(ctx, nil, createInput)
	if err != nil {
		t.Fatalf("handleCreateEvent failed: %v", err)
	}

	eventID := createRes.Event.ID
	if eventID == "" {
		t.Fatalf("Expected non-empty event ID")
	}

	// 2. List Events
	_, listRes, err := handleListEvents(ctx, nil, ListEventsInput{})
	if err != nil {
		t.Fatalf("handleListEvents failed: %v", err)
	}
	if len(listRes.Events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(listRes.Events))
	}
	if listRes.Events[0].Title != "MCPテスト飲み会" {
		t.Fatalf("Expected title 'MCPテスト飲み会', got '%s'", listRes.Events[0].Title)
	}

	// 3. Get Event
	_, getRes, err := handleGetEvent(ctx, nil, GetEventInput{EventID: eventID})
	if err != nil {
		t.Fatalf("handleGetEvent failed: %v", err)
	}
	if len(getRes.Event.Candidates) != 2 {
		t.Fatalf("Expected 2 candidates, got %d", len(getRes.Event.Candidates))
	}

	// 4. Update Event
	_, updateRes, err := handleUpdateEvent(ctx, nil, UpdateEventInput{
		EventID:     eventID,
		Title:       "MCPテスト飲み会 (更新版)",
		Description: "更新されました",
	})
	if err != nil {
		t.Fatalf("handleUpdateEvent failed: %v", err)
	}
	if updateRes.Event.Title != "MCPテスト飲み会 (更新版)" {
		t.Fatalf("Expected updated title, got '%s'", updateRes.Event.Title)
	}

	// 5. Delete Event
	_, deleteRes, err := handleDeleteEvent(ctx, nil, DeleteEventInput{EventID: eventID})
	if err != nil {
		t.Fatalf("handleDeleteEvent failed: %v", err)
	}
	if !deleteRes.Success {
		t.Fatalf("Expected delete success")
	}

	// Verify Deleted
	_, listRes2, err := handleListEvents(ctx, nil, ListEventsInput{})
	if err != nil {
		t.Fatalf("handleListEvents after delete failed: %v", err)
	}
	if len(listRes2.Events) != 0 {
		t.Fatalf("Expected 0 events after delete, got %d", len(listRes2.Events))
	}
}
