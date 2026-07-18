package handler

import (
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/sweetfish329/kanji-chan/backend/internal/database"
	"github.com/sweetfish329/kanji-chan/backend/internal/model"
)

// isSocialBot はリクエストのUser-AgentがSNS・メッセージングのOGPクローラかどうかを判定する
func isSocialBot(ua string) bool {
	ua = strings.ToLower(ua)
	bots := []string{
		"twitterbot",
		"discordbot",
		"facebookexternalhit",
		"linespider",
		"slackbot",
		"skypeuripreview",
		"telegrambot",
		"whatsapp",
		"applebot",
		"googlebot",
		"bingbot",
		"embedly",
		"vkshare",
		"outbrain",
		"pinterest",
		"bufferbot",
	}
	for _, bot := range bots {
		if strings.Contains(ua, bot) {
			return true
		}
	}
	return false
}

// eventPathRegex は /event/{uuid} のパスを検出する
var eventPathRegex = regexp.MustCompile(`^/event/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$`)

// HandleOGP はSNSボットに対してイベントのOGPメタタグ付きHTMLを返す。
// 通常ユーザーには false を返し、SPAの index.html にフォールスルーさせる。
func HandleOGP(c *echo.Context, siteURL string) bool {
	ua := c.Request().Header.Get("User-Agent")
	path := c.Request().URL.Path

	if !isSocialBot(ua) {
		return false
	}

	matches := eventPathRegex.FindStringSubmatch(path)
	if matches == nil {
		return false
	}
	eventID := matches[1]

	// DBからイベント取得
	var event model.Event
	if err := database.DB.
		Preload("Candidates").
		Where("id = ?", eventID).
		First(&event).Error; err != nil {
		return false
	}

	pageURL := fmt.Sprintf("%s/event/%s", strings.TrimRight(siteURL, "/"), eventID)
	ogpHTML := buildOGPHTML(event, pageURL, siteURL)

	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, ogpHTML)
	return true
}

// HandleOGPImage はイベントタイトル入りの動的OGP画像(SVG)を生成して返す
func HandleOGPImage(c *echo.Context) error {
	eventID := c.Param("id")
	// .png や .svg 拡張子の切り落とし
	eventID = strings.TrimSuffix(eventID, ".png")
	eventID = strings.TrimSuffix(eventID, ".svg")

	var event model.Event
	if err := database.DB.
		Preload("Candidates").
		Where("id = ?", eventID).
		First(&event).Error; err != nil {
		// 見つからない場合はデフォルトのタイトルで生成
		event = model.Event{
			Title:       "日程調整イベント",
			Description: "幹事ちゃんで日程を調整しましょう",
		}
	}

	svgContent := generateOGPSVG(event)
	c.Response().Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	return c.String(http.StatusOK, svgContent)
}

// buildOGPHTML はイベント情報からOGPメタタグを含むHTMLを生成する
func buildOGPHTML(event model.Event, pageURL, siteURL string) string {
	title := html.EscapeString(event.Title)
	siteName := "幹事ちゃん"

	descParts := []string{}
	if event.Description != "" {
		descParts = append(descParts, html.EscapeString(event.Description))
	}
	if len(event.Candidates) > 0 {
		dates := []string{}
		for i, cand := range event.Candidates {
			if i >= 3 {
				dates = append(dates, fmt.Sprintf("他%d件", len(event.Candidates)-3))
				break
			}
			parsed, err := time.Parse("2006-01-02", cand.EventDate)
			if err == nil {
				dates = append(dates, parsed.Format("1/2"))
			} else {
				dates = append(dates, cand.EventDate)
			}
		}
		descParts = append(descParts, "📅 "+strings.Join(dates, " / "))
	}
	descParts = append(descParts, "幹事ちゃんで日程を回答してください。")
	description := html.EscapeString(strings.Join(descParts, " ｜ "))

	ogImageURL := fmt.Sprintf("%s/api/ogp/%s.png", strings.TrimRight(siteURL, "/"), event.ID)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="ja" prefix="og: https://ogp.me/ns#">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>%s | %s</title>

  <!-- OGP (Open Graph Protocol) -->
  <meta property="og:type" content="website" />
  <meta property="og:url" content="%s" />
  <meta property="og:title" content="%s | %s" />
  <meta property="og:description" content="%s" />
  <meta property="og:site_name" content="%s" />
  <meta property="og:locale" content="ja_JP" />
  <meta property="og:image" content="%s" />
  <meta property="og:image:width" content="1200" />
  <meta property="og:image:height" content="630" />

  <!-- Twitter Card -->
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content="%s | %s" />
  <meta name="twitter:description" content="%s" />
  <meta name="twitter:image" content="%s" />

  <!-- noindex: イベントページは検索エンジンに表示しない -->
  <meta name="robots" content="noindex, nofollow" />

  <!-- ユーザーのブラウザは SPA にリダイレクト -->
  <meta http-equiv="refresh" content="0; url=%s" />
  <script>window.location.replace('%s');</script>
</head>
<body>
  <p><a href="%s">%s の調整ページを開く</a></p>
</body>
</html>`,
		title, siteName,
		html.EscapeString(pageURL),
		title, siteName,
		description,
		siteName,
		html.EscapeString(ogImageURL),
		title, siteName,
		description,
		html.EscapeString(ogImageURL),
		html.EscapeString(pageURL),
		html.EscapeString(pageURL),
		html.EscapeString(pageURL),
		title,
	)
}

// generateOGPSVG はイベント情報を元に美しいOGP画像(SVG 1200x630)を生成する
func generateOGPSVG(event model.Event) string {
	escapedTitle := html.EscapeString(event.Title)
	if len([]rune(escapedTitle)) > 35 {
		escapedTitle = string([]rune(escapedTitle)[:33]) + "..."
	}

	escapedDesc := html.EscapeString(event.Description)
	if len([]rune(escapedDesc)) > 60 {
		escapedDesc = string([]rune(escapedDesc)[:58]) + "..."
	}

	// 候補日時サマリー
	candidateStr := ""
	if len(event.Candidates) > 0 {
		dates := []string{}
		for i, cand := range event.Candidates {
			if i >= 4 {
				dates = append(dates, fmt.Sprintf("＋他%d件", len(event.Candidates)-4))
				break
			}
			parsed, err := time.Parse("2006-01-02", cand.EventDate)
			dateText := cand.EventDate
			if err == nil {
				dateText = parsed.Format("1/2")
			}
			if cand.StartTime != "" {
				dateText += fmt.Sprintf(" %s〜", cand.StartTime)
			}
			dates = append(dates, dateText)
		}
		candidateStr = "📅 候補日: " + strings.Join(dates, "  |  ")
	} else {
		candidateStr = "📅 〇・△・× でカンタン回答"
	}
	escapedCandidate := html.EscapeString(candidateStr)

	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1200 630" width="1200" height="630">
  <defs>
    <linearGradient id="bg-grad" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
      <stop offset="0%%" stop-color="#FAF8F5" />
      <stop offset="100%%" stop-color="#F3ECE3" />
    </linearGradient>
    <linearGradient id="card-grad" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
      <stop offset="0%%" stop-color="#FFFFFF" stop-opacity="0.9" />
      <stop offset="100%%" stop-color="#FAF8F5" stop-opacity="0.8" />
    </linearGradient>
    <linearGradient id="accent-grad" x1="0%%" y1="0%%" x2="100%%" y2="0%%">
      <stop offset="0%%" stop-color="#8A7463" />
      <stop offset="100%%" stop-color="#C49B82" />
    </linearGradient>
    <filter id="shadow" x="-5%%" y="-5%%" width="110%%" height="110%%">
      <feDropShadow dx="0" dy="16" stdDeviation="24" flood-color="#2C2621" flood-opacity="0.08" />
    </filter>
  </defs>

  <!-- 背景 -->
  <rect width="1200" height="630" fill="url(#bg-grad)" />

  <!-- 装飾用のグラデーション円 -->
  <circle cx="1050" cy="100" r="280" fill="#5E6F62" opacity="0.06" />
  <circle cx="150" cy="530" r="220" fill="#C49B82" opacity="0.08" />

  <!-- メインカード -->
  <rect x="80" y="70" width="1040" height="490" rx="32" fill="url(#card-grad)" stroke="rgba(94, 83, 74, 0.12)" stroke-width="2" filter="url(#shadow)" />

  <!-- ブランドロゴ header -->
  <g transform="translate(140, 135)">
    <rect x="0" y="0" width="40" height="40" rx="12" fill="#5E6F62" />
    <path d="M12 10 v20 M28 10 v20 M8 18 h24 M8 26 h24" stroke="#FAF8F5" stroke-width="3" stroke-linecap="round" />
    <text x="56" y="28" font-family="Plus Jakarta Sans, -apple-system, sans-serif" font-weight="700" font-size="24" fill="#8A7463" letter-spacing="1">幹事ちゃん</text>
    <text x="175" y="28" font-family="Plus Jakarta Sans, -apple-system, sans-serif" font-size="16" fill="#93857B">AIサポート日程調整</text>
  </g>

  <!-- アクセントライン -->
  <rect x="140" y="195" width="92" height="4" rx="2" fill="url(#accent-grad)" />

  <!-- イベントタイトル -->
  <text x="140" y="275" font-family="'Plus Jakarta Sans', 'Hiragino Sans', 'Meiryo', sans-serif" font-weight="700" font-size="44" fill="#2C2621">%s</text>

  <!-- 説明文 (あれば) -->
  <text x="140" y="340" font-family="'Plus Jakarta Sans', 'Hiragino Sans', 'Meiryo', sans-serif" font-size="22" fill="#5E534A">%s</text>

  <!-- 候補日ピル -->
  <g transform="translate(140, 400)">
    <rect x="0" y="0" width="820" height="60" rx="20" fill="#F3ECE3" opacity="0.7" stroke="rgba(94,83,74,0.1)" stroke-width="1" />
    <text x="24" y="37" font-family="'Plus Jakarta Sans', 'Hiragino Sans', 'Meiryo', sans-serif" font-weight="600" font-size="20" fill="#2C2621">%s</text>
  </g>

  <!-- フッターのアクション誘導 -->
  <g transform="translate(140, 505)">
    <text x="0" y="0" font-family="'Plus Jakarta Sans', 'Hiragino Sans', 'Meiryo', sans-serif" font-weight="600" font-size="18" fill="#5E6F62">✨ 登録・ログイン不要で〇△×入力できます</text>
  </g>
</svg>`,
		escapedTitle,
		escapedDesc,
		escapedCandidate,
	)
}
