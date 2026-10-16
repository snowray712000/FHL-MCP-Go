package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SmokeResult is what a successful smoke test saw.
type SmokeResult struct {
	ServerName      string
	ServerVersion   string
	ProtocolVersion string
	ToolCount       int
}

// Smoke runs a real MCP session against url: initialize, tools/list and one representative call
// (get_book_list, which is static data and never reaches the FHL upstream).
func Smoke(ctx context.Context, url string) (SmokeResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	c := mcp.NewClient(&mcp.Implementation{Name: "fhl-mcp-smoke", Version: "1"}, nil)
	sess, err := c.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: url}, nil)
	if err != nil {
		return SmokeResult{}, fmt.Errorf("initialize failed: %w", err)
	}
	defer sess.Close()
	init := sess.InitializeResult()
	res := SmokeResult{ServerName: init.ServerInfo.Name, ServerVersion: init.ServerInfo.Version, ProtocolVersion: init.ProtocolVersion}

	tl, err := sess.ListTools(ctx, nil)
	if err != nil {
		return res, fmt.Errorf("tools/list failed: %w", err)
	}
	res.ToolCount = len(tl.Tools)

	call, err := sess.CallTool(ctx, &mcp.CallToolParams{Name: "get_book_list", Arguments: map[string]any{"testament": "NT"}})
	if err != nil {
		return res, fmt.Errorf("tools/call get_book_list failed: %w", err)
	}
	if call.IsError {
		return res, fmt.Errorf("get_book_list returned an error result")
	}
	var text strings.Builder
	for _, ct := range call.Content {
		if t, ok := ct.(*mcp.TextContent); ok {
			text.WriteString(t.Text)
		}
	}
	if !strings.Contains(text.String(), "Matthew") && !strings.Contains(text.String(), "馬太") {
		return res, fmt.Errorf("get_book_list returned unexpected content")
	}
	return res, nil
}
