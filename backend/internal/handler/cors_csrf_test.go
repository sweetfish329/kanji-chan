package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func TestCORSOriginWhitelist(t *testing.T) {
	e := echo.New()

	allowedOrigins := map[string]bool{
		"http://localhost:5173": true,
		"http://localhost:8080": true,
	}

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		UnsafeAllowOriginFunc: func(c *echo.Context, origin string) (string, bool, error) {
			cleanOrigin := strings.TrimRight(origin, "/")
			if allowedOrigins[cleanOrigin] {
				return origin, true, nil
			}
			return "", false, nil
		},
		AllowCredentials: true,
	}))

	e.GET("/api/test-cors", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Test 1: Valid Origin
	reqValid := httptest.NewRequest(http.MethodOptions, "/api/test-cors", nil)
	reqValid.Header.Set("Origin", "http://localhost:5173")
	recValid := httptest.NewRecorder()
	e.ServeHTTP(recValid, reqValid)

	if recValid.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Errorf("Expected Access-Control-Allow-Origin to be http://localhost:5173, got %q", recValid.Header().Get("Access-Control-Allow-Origin"))
	}

	// Test 2: Invalid/Malicious Origin
	reqInvalid := httptest.NewRequest(http.MethodOptions, "/api/test-cors", nil)
	reqInvalid.Header.Set("Origin", "http://evil-attacker.com")
	recInvalid := httptest.NewRecorder()
	e.ServeHTTP(recInvalid, reqInvalid)

	if recInvalid.Header().Get("Access-Control-Allow-Origin") == "http://evil-attacker.com" {
		t.Errorf("Expected Access-Control-Allow-Origin to NOT allow malicious origin, but got allowed")
	}
}

func TestCSRFProtectionMiddleware(t *testing.T) {
	e := echo.New()

	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "header:X-CSRF-Token,form:_csrf",
		CookieName:     "_csrf",
		CookiePath:     "/",
		CookieHTTPOnly: false,
		CookieSameSite: http.SameSiteLaxMode,
		Skipper: func(c *echo.Context) bool {
			if c.Request().Method == http.MethodGet || c.Request().Method == http.MethodOptions {
				return true
			}
			if apiKey := c.Request().Header.Get("X-API-Key"); apiKey != "" {
				return true
			}
			return false
		},
	}))

	e.POST("/api/test-csrf", func(c *echo.Context) error {
		return c.String(http.StatusOK, "POST Success")
	})

	// Test 1: POST without CSRF token should be rejected (403 Forbidden)
	reqNoCSRF := httptest.NewRequest(http.MethodPost, "/api/test-csrf", nil)
	recNoCSRF := httptest.NewRecorder()
	e.ServeHTTP(recNoCSRF, reqNoCSRF)

	if recNoCSRF.Code != http.StatusBadRequest && recNoCSRF.Code != http.StatusForbidden {
		t.Errorf("Expected POST without CSRF token to be rejected (400/403), got %d", recNoCSRF.Code)
	}

	// Test 2: POST with API Key header should skip CSRF check
	reqAPIKey := httptest.NewRequest(http.MethodPost, "/api/test-csrf", nil)
	reqAPIKey.Header.Set("X-API-Key", "kc_valid_api_key")
	recAPIKey := httptest.NewRecorder()
	e.ServeHTTP(recAPIKey, reqAPIKey)

	if recAPIKey.Code != http.StatusOK {
		t.Errorf("Expected POST with API Key to succeed (200), got %d", recAPIKey.Code)
	}
}
