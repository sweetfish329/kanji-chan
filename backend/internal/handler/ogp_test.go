package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/sweetfish329/kanji-chan/backend/internal/model"
)

func TestIsSocialBot(t *testing.T) {
	tests := []struct {
		ua       string
		expected bool
	}{
		{"Mozilla/5.0 (compatible; Twitterbot/1.0)", true},
		{"facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)", true},
		{"Mozilla/5.0 (compatible; Discordbot/2.0; +https://discordapp.com)", true},
		{"Linespider-implicit/1.0", true},
		{"Slackbot-LinkExpanding 1.0 (+https://api.slack.com/robots)", true},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36", false},
	}

	for _, tt := range tests {
		result := isSocialBot(tt.ua)
		if result != tt.expected {
			t.Errorf("isSocialBot(%q) = %v; want %v", tt.ua, result, tt.expected)
		}
	}
}

func TestGetSiteURL(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "http://kanji-chan.example.com/event/123", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "kanji-chan.example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	siteURL := GetSiteURL(c, "http://localhost:8080")
	expected := "https://kanji-chan.example.com"

	if siteURL != expected {
		t.Errorf("GetSiteURL() = %q; want %q", siteURL, expected)
	}
}

func TestGenerateOGPSVG(t *testing.T) {
	event := model.Event{
		ID:          "12345678-1234-1234-1234-1234567890ab",
		Title:       "テストイベント飲み会",
		Description: "渋谷でみんなで楽しく飲みましょう",
		Candidates: []model.EventCandidate{
			{EventDate: "2026-07-20", StartTime: "19:00", EndTime: "21:00"},
		},
	}

	svg := generateOGPSVG(event)
	if svg == "" {
		t.Fatal("generateOGPSVG returned empty string")
	}

	if !testing.Short() {
		if !containsString(svg, "テストイベント飲み会") {
			t.Errorf("SVG does not contain title")
		}
		if !containsString(svg, "幹事ちゃん") {
			t.Errorf("SVG does not contain brand name")
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
