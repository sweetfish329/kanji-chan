package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sweetfish329/kanji-chan/backend/internal/auth"
	"github.com/sweetfish329/kanji-chan/backend/internal/database"
	"github.com/sweetfish329/kanji-chan/backend/internal/handler"
)

func main() {
	// ローカル開発時は .env ファイルをロード (コンテナ時は環境変数が直接渡されるため無視してOK)
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// データベース接続の初期化
	_, err := database.InitDB()
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	// 認証の初期設定
	auth.InitAuth()

	e := echo.New()

	// ミドルウェアの設定
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOriginFunc: func(origin string) (bool, error) {
			// クッキー認証（Credentials: true）と任意のOrigin許可を両立するための動的Origin判定
			return true, nil
		},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAuthorization, "X-Response-Token"},
		AllowCredentials: true,
	}))

	// 共通・認証 (パブリック)
	e.GET("/api/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "ok",
			"message": "Kanji-Chan API is running",
		})
	})
	e.GET("/api/auth/login", handler.HandleLogin)
	e.GET("/api/auth/callback", handler.HandleCallback)
	e.POST("/api/auth/logout", handler.HandleLogout)

	// イベント詳細・回答登録 (パブリック)
	e.GET("/api/events/:id", handler.HandleGetEvent)
	e.POST("/api/events/:id/responses", handler.HandleAddResponse)
	e.PUT("/api/events/:id/responses/:response_id", handler.HandleUpdateResponse)
	e.DELETE("/api/events/:id/responses/:response_id", handler.HandleDeleteResponse)

	// 認証が必要なプライベートグループ
	r := e.Group("")
	r.Use(handler.AuthMiddleware)

	r.GET("/api/auth/me", handler.HandleMe)
	r.POST("/api/auth/apikey", handler.HandleUpdateAPIKey)
	r.POST("/api/events", handler.HandleCreateEvent)
	r.GET("/api/events", handler.HandleListEvents)
	r.PUT("/api/events/:id", handler.HandleUpdateEvent)
	r.DELETE("/api/events/:id", handler.HandleDeleteEvent)
	r.POST("/api/ai/parse-event", handler.HandleParseEvent)
	r.POST("/api/ai/suggest-schedule", handler.HandleSuggestSchedule)

	log.Printf("Starting Kanji-Chan backend server on port %s...", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
