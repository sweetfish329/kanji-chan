package handler

import (
	"net/http"
	"testing"

	"github.com/sweetfish329/kanji-chan/backend/internal/auth"
	"github.com/sweetfish329/kanji-chan/backend/internal/model"
)

func TestValidateEventManagementAccess(t *testing.T) {
	creatorID := uint(1)
	otherUserID := uint(2)

	event := &model.Event{
		ID:        "test-event-uuid",
		Title:     "テストイベント",
		CreatedBy: &creatorID,
	}

	tests := []struct {
		name       string
		claims     *auth.Claims
		shouldFail bool
		statusCode int
	}{
		{
			name: "Owner user access",
			claims: &auth.Claims{
				UserID:   creatorID,
				Role:     "user",
				IsAPIKey: false,
			},
			shouldFail: false,
		},
		{
			name: "Admin user access (different ID)",
			claims: &auth.Claims{
				UserID:   otherUserID,
				Role:     "admin",
				IsAPIKey: false,
			},
			shouldFail: false,
		},
		{
			name: "Admin API Key access (owner)",
			claims: &auth.Claims{
				UserID:   creatorID,
				Role:     "admin",
				IsAPIKey: true,
			},
			shouldFail: false,
		},
		{
			name: "Non-admin API Key access (BOLA/Privilege Escalation attempt)",
			claims: &auth.Claims{
				UserID:   creatorID,
				Role:     "user",
				IsAPIKey: true,
			},
			shouldFail: true,
			statusCode: http.StatusForbidden,
		},
		{
			name: "Other non-admin user access (BOLA attempt)",
			claims: &auth.Claims{
				UserID:   otherUserID,
				Role:     "user",
				IsAPIKey: false,
			},
			shouldFail: true,
			statusCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEventManagementAccess(tt.claims, event)
			if tt.shouldFail {
				if err == nil {
					t.Errorf("expected error for %s, got nil", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for %s: %v", tt.name, err)
				}
			}
		})
	}
}
