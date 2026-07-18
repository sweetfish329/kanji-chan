package mcp

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sweetfish329/kanji-chan/backend/internal/auth"
	"github.com/sweetfish329/kanji-chan/backend/internal/handler"
)

type contextKey string

const UserClaimsKey contextKey = "user_claims"

// GetClaimsFromContext MCPツールの context.Context から認証ユーザー情報を取得
func GetClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(*auth.Claims)
	return claims, ok
}

// BuildMCPServer MCPサーバーの初期化とツールの登録
func BuildMCPServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "kanji-chan-mcp",
		Version: "1.0.0",
	}, nil)

	RegisterTools(server)

	return server
}

// NewHandler Echo用のMCPハンドラー生成 (Streamable HTTP)
func NewHandler() echo.HandlerFunc {
	mcpServer := BuildMCPServer()

	streamableHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return mcpServer
	}, nil)

	return func(c *echo.Context) error {
		// Echo コンテキスト内の User claims を http.Request Context に注入
		req := c.Request()
		if claims, ok := handler.GetUserFromContext(c); ok {
			ctx := context.WithValue(req.Context(), UserClaimsKey, claims)
			req = req.WithContext(ctx)
		}

		streamableHandler.ServeHTTP(c.Response(), req)
		return nil
	}
}
