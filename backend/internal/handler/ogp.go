package handler

import (
	"context"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/sweetfish329/kanji-chan/backend/internal/database"
	"github.com/sweetfish329/kanji-chan/backend/internal/model"
)

// --- SSRF (Server-Side Request Forgery) 防御層 & OGP取得処理 ---

// isPrivateOrReservedIP 指定されたIPアドレスがプライベート/ループバック/リンクローカル/保留IPでないかをチェック
func isPrivateOrReservedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	// ループバック (127.0.0.0/8, ::1)
	if ip.IsLoopback() {
		return true
	}
	// プライベートIP (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, fc00::/7)
	if ip.IsPrivate() {
		return true
	}
	// リンクローカルユニキャスト/マルチキャスト (169.254.0.0/16 AWS IMDS 含む, fe80::/10)
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	// 未指定IP (0.0.0.0, ::)
	if ip.IsUnspecified() {
		return true
	}
	// AWSメタデータIP (169.254.169.254) およびマルチキャストIP
	if ip.Equal(net.ParseIP("169.254.169.254")) || ip.IsMulticast() {
		return true
	}
	return false
}

// validateURLForSSRF 入力されたURLのスキーム制限およびDNS名前解決後のIPアドレスバリデーション
func validateURLForSSRF(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("prohibited scheme: %s (only http and https are allowed)", scheme)
	}

	hostname := parsed.Hostname()
	if hostname == "" {
		return fmt.Errorf("empty hostname")
	}

	// ホスト名のDNS名前解決を行い、解決されたすべてのIPをテスト
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return fmt.Errorf("failed to resolve hostname %s: %w", hostname, err)
	}

	for _, ip := range ips {
		if isPrivateOrReservedIP(ip) {
			return fmt.Errorf("access to private or internal IP address %s is prohibited", ip.String())
		}
	}

	return nil
}

// safeHTTPClient DNS Dynamic Rebinding (TOCTOU) を防止するためのソケット接続レベルでのIP判定ダイアラー
var safeHTTPClient = &http.Client{
	Timeout: 5 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("too many redirects")
		}
		// リダイレクト先のURLも再度SSRFチェック
		return validateURLForSSRF(req.URL.String())
	},
	Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
			if err != nil {
				return nil, err
			}
			for _, ip := range ips {
				if isPrivateOrReservedIP(ip) {
					return nil, fmt.Errorf("access to private/reserved IP address %s is prohibited", ip.String())
				}
			}
			dialer := &net.Dialer{
				Timeout: 3 * time.Second,
			}
			return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].String(), port))
		},
	},
}

type OGPFetchResponse struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	URL         string `json:"url"`
}

// HandleFetchOGP 外部URLからOGP情報を安全に取得する (SSRF対策済み)
func HandleFetchOGP(c *echo.Context) error {
	targetURL := c.QueryParam("url")
	if targetURL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing 'url' query parameter")
	}

	// スキーム制限 (http/httpsのみ) および プライベートIP拒否バリデーション
	if err := validateURLForSSRF(targetURL); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "SSRF validation failed: "+err.Error())
	}

	req, err := http.NewRequestWithContext(c.Request().Context(), http.MethodGet, targetURL, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to create HTTP request")
	}
	req.Header.Set("User-Agent", "KanjiChan-OGPFetcher/1.0")

	resp, err := safeHTTPClient.Do(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to fetch target URL: "+err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Target URL returned HTTP status %d", resp.StatusCode))
	}

	// 読み込み制限 (1MB)
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to read response body")
	}
	bodyStr := string(bodyBytes)

	ogpData := OGPFetchResponse{
		URL: targetURL,
	}

	if matches := regexp.MustCompile(`(?i)<meta\s+property=["']og:title["']\s+content=["']([^"']+)["']`).FindStringSubmatch(bodyStr); len(matches) > 1 {
		ogpData.Title = html.UnescapeString(matches[1])
	} else if matches := regexp.MustCompile(`(?i)<title>([^<]+)</title>`).FindStringSubmatch(bodyStr); len(matches) > 1 {
		ogpData.Title = html.UnescapeString(matches[1])
	}

	if matches := regexp.MustCompile(`(?i)<meta\s+property=["']og:description["']\s+content=["']([^"']+)["']`).FindStringSubmatch(bodyStr); len(matches) > 1 {
		ogpData.Description = html.UnescapeString(matches[1])
	} else if matches := regexp.MustCompile(`(?i)<meta\s+name=["']description["']\s+content=["']([^"']+)["']`).FindStringSubmatch(bodyStr); len(matches) > 1 {
		ogpData.Description = html.UnescapeString(matches[1])
	}

	if matches := regexp.MustCompile(`(?i)<meta\s+property=["']og:image["']\s+content=["']([^"']+)["']`).FindStringSubmatch(bodyStr); len(matches) > 1 {
		ogpData.Image = html.UnescapeString(matches[1])
	}

	return c.JSON(http.StatusOK, ogpData)
}

// isSocialBot はリクエストのUser-AgentがSNS・メッセージングのOGPクローラかどうかを判定する
func isSocialBot(ua string) bool {
	ua = strings.ToLower(ua)
	bots := []string{
		"twitterbot",
		"xbot",
		"facebookexternalhit",
		"facebot",
		"meta-externalagent",
		"linespider",
		"line-poker",
		"discordbot",
		"slackbot",
		"slack-imgproxy",
		"telegrambot",
		"whatsapp",
		"skypeuripreview",
		"applebot",
		"googlebot",
		"bingbot",
		"embedly",
		"vkshare",
		"outbrain",
		"pinterest",
		"hatena",
		"yandex",
		"curl",
		"wget",
	}
	for _, bot := range bots {
		if strings.Contains(ua, bot) {
			return true
		}
	}
	return false
}

// GetSiteURL はリクエストヘッダーまたは環境変数から正しい絶対URLを生成する
func GetSiteURL(c *echo.Context, defaultSiteURL string) string {
	if envURL := os.Getenv("PUBLIC_SITE_URL"); envURL != "" {
		return strings.TrimRight(envURL, "/")
	}

	if c == nil || c.Request() == nil {
		if defaultSiteURL != "" {
			return strings.TrimRight(defaultSiteURL, "/")
		}
		return "http://localhost:8080"
	}

	scheme := "http"
	if proto := c.Request().Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	} else if c.Request().TLS != nil {
		scheme = "https"
	}

	host := c.Request().Header.Get("X-Forwarded-Host")
	if host == "" {
		host = c.Request().Host
	}

	if host != "" {
		return fmt.Sprintf("%s://%s", scheme, host)
	}

	if defaultSiteURL != "" {
		return strings.TrimRight(defaultSiteURL, "/")
	}
	return "http://localhost:8080"
}

// eventPathRegex は /event/{uuid} または /event/{uuid}/ などのパスを検出する (大文字小文字対応)
var eventPathRegex = regexp.MustCompile(`(?i)^/event/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})/?$`)

// HandleOGP はSNSボットに対してイベントのOGPメタタグ付きHTMLを返す。
// 通常ユーザーには false を返し、SPAの index.html にフォールスルーさせる。
func HandleOGP(c *echo.Context, defaultSiteURL string) bool {
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

	siteURL := GetSiteURL(c, defaultSiteURL)
	pageURL := fmt.Sprintf("%s/event/%s", siteURL, eventID)
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
	c.Response().Header().Set("Cache-Control", "public, max-age=86400, s-maxage=86400")
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
  <meta property="og:image:type" content="image/svg+xml" />
  <meta property="og:image:width" content="1200" />
  <meta property="og:image:height" content="630" />

  <!-- Twitter Card -->
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content="%s | %s" />
  <meta name="twitter:description" content="%s" />
  <meta name="twitter:image" content="%s" />

  <!-- noindex: イベントページは検索エンジンインデックスから除外 -->
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

// generateOGPSVG はイベント情報を元に和モダンスタイルの美しく視認性の高いOGP画像(SVG 1200x630)を生成する
func generateOGPSVG(event model.Event) string {
	escapedTitle := html.EscapeString(event.Title)
	runesTitle := []rune(escapedTitle)

	// タイトルの自動折り返し・切り詰め
	titleLine1 := string(runesTitle)
	titleLine2 := ""
	if len(runesTitle) > 18 {
		titleLine1 = string(runesTitle[:18])
		remaining := runesTitle[18:]
		if len(remaining) > 18 {
			titleLine2 = string(remaining[:16]) + "..."
		} else {
			titleLine2 = string(remaining)
		}
	}

	escapedDesc := html.EscapeString(event.Description)
	if len([]rune(escapedDesc)) > 50 {
		escapedDesc = string([]rune(escapedDesc)[:48]) + "..."
	}

	// 候補日時サマリー
	candidateStr := ""
	if len(event.Candidates) > 0 {
		dates := []string{}
		for i, cand := range event.Candidates {
			if i >= 3 {
				dates = append(dates, fmt.Sprintf("他%d件", len(event.Candidates)-3))
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

	titleSVG := fmt.Sprintf(`<text x="140" y="275" font-family="'Zen Kaku Gothic New', 'Plus Jakarta Sans', 'Hiragino Sans', 'Meiryo', sans-serif" font-weight="800" font-size="44" fill="#1C241E">%s</text>`, titleLine1)
	if titleLine2 != "" {
		titleSVG = fmt.Sprintf(`
			<text x="140" y="260" font-family="'Zen Kaku Gothic New', 'Plus Jakarta Sans', 'Hiragino Sans', 'Meiryo', sans-serif" font-weight="800" font-size="40" fill="#1C241E">%s</text>
			<text x="140" y="315" font-family="'Zen Kaku Gothic New', 'Plus Jakarta Sans', 'Hiragino Sans', 'Meiryo', sans-serif" font-weight="800" font-size="40" fill="#1C241E">%s</text>
		`, titleLine1, titleLine2)
	}

	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1200 630" width="1200" height="630">
  <defs>
    <linearGradient id="bg-grad" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
      <stop offset="0%%" stop-color="#F8F6F0" />
      <stop offset="100%%" stop-color="#EFE8DC" />
    </linearGradient>
    <linearGradient id="card-grad" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
      <stop offset="0%%" stop-color="#FFFFFF" stop-opacity="0.95" />
      <stop offset="100%%" stop-color="#F8F6F0" stop-opacity="0.88" />
    </linearGradient>
    <linearGradient id="accent-grad" x1="0%%" y1="0%%" x2="100%%" y2="0%%">
      <stop offset="0%%" stop-color="#2A4032" />
      <stop offset="50%%" stop-color="#44614E" />
      <stop offset="100%%" stop-color="#D48C38" />
    </linearGradient>
    <filter id="shadow" x="-5%%" y="-5%%" width="110%%" height="110%%">
      <feDropShadow dx="0" dy="16" stdDeviation="24" flood-color="#1C241E" flood-opacity="0.08" />
    </filter>
  </defs>

  <!-- 背景 -->
  <rect width="1200" height="630" fill="url(#bg-grad)" />

  <!-- 和モダン装飾円 -->
  <circle cx="1050" cy="100" r="280" fill="#2A4032" opacity="0.06" />
  <circle cx="150" cy="530" r="220" fill="#D48C38" opacity="0.07" />

  <!-- メインカード -->
  <rect x="80" y="70" width="1040" height="490" rx="28" fill="url(#card-grad)" stroke="rgba(42, 64, 50, 0.1)" stroke-width="2" filter="url(#shadow)" />

  <!-- ブランドロゴ header -->
  <g transform="translate(140, 135)">
    <rect x="0" y="0" width="44" height="44" rx="12" fill="#2A4032" />
    <path d="M14 12 v20 M30 12 v20 M10 20 h24 M10 28 h24" stroke="#F8F6F0" stroke-width="3" stroke-linecap="round" />
    <text x="60" y="30" font-family="'Zen Kaku Gothic New', 'Plus Jakarta Sans', sans-serif" font-weight="800" font-size="26" fill="#2A4032" letter-spacing="1">幹事ちゃん</text>
    <text x="195" y="30" font-family="'Zen Kaku Gothic New', 'Plus Jakarta Sans', sans-serif" font-size="16" fill="#8C948B">AI スケジュール調整</text>
  </g>

  <!-- アクセントライン -->
  <rect x="140" y="198" width="120" height="4" rx="2" fill="url(#accent-grad)" />

  <!-- イベントタイトル -->
  %s

  <!-- 説明文 -->
  <text x="140" y="360" font-family="'Zen Kaku Gothic New', 'Plus Jakarta Sans', 'Hiragino Sans', 'Meiryo', sans-serif" font-size="22" fill="#4A544C">%s</text>

  <!-- 候補日ピル -->
  <g transform="translate(140, 410)">
    <rect x="0" y="0" width="820" height="60" rx="16" fill="#EFE8DC" opacity="0.8" stroke="rgba(42, 64, 50, 0.1)" stroke-width="1" />
    <text x="24" y="37" font-family="'Zen Kaku Gothic New', 'Plus Jakarta Sans', 'Hiragino Sans', 'Meiryo', sans-serif" font-weight="700" font-size="20" fill="#1C241E">%s</text>
  </g>

  <!-- フッターのアクション誘導 -->
  <g transform="translate(140, 510)">
    <text x="0" y="0" font-family="'Zen Kaku Gothic New', 'Plus Jakarta Sans', 'Hiragino Sans', 'Meiryo', sans-serif" font-weight="700" font-size="18" fill="#2A4032">✨ ログイン不要で〇・△・× 回答できます</text>
  </g>
</svg>`,
		titleSVG,
		escapedDesc,
		escapedCandidate,
	)
}
