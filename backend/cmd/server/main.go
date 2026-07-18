package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/sweetfish329/kanji-chan/backend/internal/auth"
	"github.com/sweetfish329/kanji-chan/backend/internal/database"
	"github.com/sweetfish329/kanji-chan/backend/internal/handler"
	"github.com/sweetfish329/kanji-chan/backend/internal/mcp"
)

//go:embed dist/*
var webAssets embed.FS

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
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		UnsafeAllowOriginFunc: func(c *echo.Context, origin string) (string, bool, error) {
			// クッキー認証（Credentials: true）と任意のOrigin許可を両立するための動的Origin判定
			return origin, true, nil
		},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAuthorization, "X-Response-Token"},
		AllowCredentials: true,
	}))

	// 共通・認証 (パブリック)
	e.GET("/api/health", func(c *echo.Context) error {
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

	// OGP動的画像生成 (パブリック)
	e.GET("/api/ogp/:id", handler.HandleOGPImage)

	// 認証が必要なプライベートグループ
	r := e.Group("")

	r.GET("/api/auth/me", handler.HandleMe, handler.AuthMiddleware)
	r.POST("/api/auth/apikey", handler.HandleUpdateAPIKey, handler.AuthMiddleware)
	r.GET("/api/auth/apikeys", handler.HandleListAPIKeys, handler.AuthMiddleware)
	r.POST("/api/auth/apikeys", handler.HandleCreateAPIKey, handler.AuthMiddleware)
	r.DELETE("/api/auth/apikeys/:id", handler.HandleDeleteAPIKey, handler.AuthMiddleware)
	r.POST("/api/events", handler.HandleCreateEvent, handler.AuthMiddleware)
	r.GET("/api/events", handler.HandleListEvents, handler.AuthMiddleware)
	r.PUT("/api/events/:id", handler.HandleUpdateEvent, handler.AuthMiddleware)
	r.DELETE("/api/events/:id", handler.HandleDeleteEvent, handler.AuthMiddleware)
	r.POST("/api/ai/parse-event", handler.HandleParseEvent, handler.AuthMiddleware)
	r.POST("/api/ai/suggest-schedule", handler.HandleSuggestSchedule, handler.AuthMiddleware)

	// Streamable HTTP MCP (Model Context Protocol) ルート (エンドポイント: /mcp)
	e.Any("/mcp", mcp.NewHandler(), handler.AuthMiddleware)

	// ==========================================
	// SEO: robots.txt / sitemap.xml を動的生成
	// PUBLIC_SITE_URL 環境変数でサイトURLを制御
	// 例: PUBLIC_SITE_URL=https://example.com
	// ==========================================
	siteURL := os.Getenv("PUBLIC_SITE_URL")
	if siteURL == "" {
		siteURL = "http://localhost:" + port
	}
	// 末尾スラッシュを除去
	siteURL = strings.TrimRight(siteURL, "/")

	e.GET("/robots.txt", func(c *echo.Context) error {
		c.Response().Header().Set("Content-Type", "text/plain; charset=utf-8")
		body := fmt.Sprintf(`User-agent: *
Allow: /$
Disallow: /event/
Disallow: /admin/
Disallow: /api/

# SNS・メッセージングボット向けOGPプレビューの許可
User-agent: Twitterbot
User-agent: Discordbot
User-agent: facebookexternalhit
User-agent: Linespider
User-agent: Slackbot
User-agent: SkypeUriPreview
Allow: /event/
Allow: /api/ogp/
Disallow: /admin/

Sitemap: %s/sitemap.xml
`, siteURL)
		return c.String(http.StatusOK, body)
	})

	e.GET("/sitemap.xml", func(c *echo.Context) error {
		c.Response().Header().Set("Content-Type", "application/xml; charset=utf-8")
		body := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"
        xmlns:xhtml="http://www.w3.org/1999/xhtml">
  <url>
    <loc>%s/</loc>
    <changefreq>weekly</changefreq>
    <priority>1.0</priority>
    <xhtml:link rel="alternate" hreflang="ja" href="%s/"/>
  </url>
</urlset>
`, siteURL, siteURL)
		return c.String(http.StatusOK, body)
	})

	// フロントエンドの静的アセット配信 (SPAルーティング対応)
	assetFS, err := fs.Sub(webAssets, "dist")
	if err != nil {
		log.Fatalf("Failed to create static asset filesystem: %v", err)
	}

	staticHandler := echo.StaticDirectoryHandler(assetFS, false)

	e.GET("/*", func(c *echo.Context) error {
		path := c.Param("*")
		if path == "" {
			path = "index.html"
		}

		// SNSボットからのリクエストで /event/ の場合は OGP HTML を返す
		if handler.HandleOGP(c, siteURL) {
			return nil
		}

		// ==========================================
		// SEO: X-Robots-Tag HTTPヘッダーの設定
		// ルートページのみ index, それ以外の HTML ページは noindex
		// (静的アセット・robots.txt・sitemap.xml は除外)
		// ==========================================
		reqPath := c.Request().URL.Path
		isStaticAsset := len(reqPath) > 1 && (reqPath[1] == '_' ||
			reqPath == "/robots.txt" ||
			reqPath == "/sitemap.xml" ||
			reqPath == "/favicon.ico" ||
			reqPath == "/favicon.svg")
		if !isStaticAsset {
			if reqPath == "/" || reqPath == "" {
				c.Response().Header().Set("X-Robots-Tag", "index, follow")
			} else {
				// /event/*, /admin/* はクローラーに非表示
				c.Response().Header().Set("X-Robots-Tag", "noindex, nofollow")
			}
		}

		// ファイルが存在するかチェック
		file, err := assetFS.Open(path)
		if err == nil {
			file.Close()
			return staticHandler(c)
		}

		// 存在しない場合は、SPAルーティングのために index.html を返す
		indexFile, err := assetFS.Open("index.html")
		if err != nil {
			return c.String(http.StatusNotFound, "Not Found")
		}
		defer indexFile.Close()

		stat, err := indexFile.Stat()
		if err != nil {
			return c.String(http.StatusInternalServerError, "Internal Server Error")
		}

		seeker, ok := indexFile.(io.ReadSeeker)
		if !ok {
			content, err := io.ReadAll(indexFile)
			if err != nil {
				return c.String(http.StatusInternalServerError, "Internal Server Error")
			}
			return c.HTML(http.StatusOK, string(content))
		}

		http.ServeContent(c.Response(), c.Request(), "index.html", stat.ModTime(), seeker)
		return nil
	})

	log.Printf("Starting Kanji-Chan backend server on port %s...", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
