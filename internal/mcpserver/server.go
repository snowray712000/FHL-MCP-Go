// Package mcpserver is the single registration source for MCP features. It exposes the frozen
// 29-tool contract and routes calls to internal/tools. Prompts and Resources are deferred to M4.
package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/snowray712000/FHL-MCP-Go/assets"
	"github.com/snowray712000/FHL-MCP-Go/internal/bible"
	"github.com/snowray712000/FHL-MCP-Go/internal/fhlapi"
	"github.com/snowray712000/FHL-MCP-Go/internal/tools"
)

// Caller runs a tool by name with raw JSON arguments.
type Caller interface {
	Call(ctx context.Context, name string, raw json.RawMessage) (string, error)
}

// Options configure New.
type Options struct {
	Name    string
	Version string
	Tools   Caller
	Logger  *slog.Logger
}

// New builds the MCP server: every tool in the embedded contract is registered verbatim
// (schemas are not regenerated), and an unknown tool name yields an isError result.
func New(o Options) (*mcp.Server, error) {
	var contract struct {
		Tools []mcp.Tool `json:"tools"`
	}
	if err := json.Unmarshal(assets.ToolsListJSON, &contract); err != nil {
		return nil, fmt.Errorf("embedded tools list: %w", err)
	}
	if len(contract.Tools) == 0 {
		return nil, errors.New("embedded tools list is empty")
	}

	srv := mcp.NewServer(&mcp.Implementation{Name: o.Name, Version: o.Version}, &mcp.ServerOptions{
		Logger: o.Logger,
		// Tools only (first release). No logging/prompts/resources capability is advertised.
		Capabilities: &mcp.ServerCapabilities{Tools: &mcp.ToolCapabilities{ListChanged: false}},
	})

	known := make(map[string]bool, len(contract.Tools))
	for i := range contract.Tools {
		t := contract.Tools[i]
		known[t.Name] = true
		srv.AddTool(&t, toolHandler(o.Tools, t.Name, o.Logger))
	}

	// D2: the SDK answers an unknown tool with a JSON-RPC -32602 error; the contract wants an
	// isError result with the same text prefix as every other tool error.
	srv.AddReceivingMiddleware(func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			if method == "tools/call" {
				if cr, ok := req.(*mcp.CallToolRequest); ok && cr.Params != nil && !known[cr.Params.Name] {
					return errorResult(cr.Params.Name, "Unknown tool: "+cr.Params.Name), nil
				}
			}
			return next(ctx, method, req)
		}
	})
	return srv, nil
}

func toolHandler(c Caller, name string, log *slog.Logger) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		text, err := c.Call(ctx, name, req.Params.Arguments)
		if err != nil {
			if log != nil {
				log.Warn("tool error", "tool", name, "err", err)
			}
			return errorResult(name, errorMessage(name, err)), nil
		}
		// Same string in both places (contract: content[0].text == structuredContent.result).
		return &mcp.CallToolResult{
			Content:           []mcp.Content{&mcp.TextContent{Text: text}},
			StructuredContent: map[string]any{"result": text},
		}, nil
	}
}

func errorResult(tool, msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: "Error executing tool " + tool + ": " + msg}},
	}
}

func errorMessage(tool string, err error) string {
	var ip *bible.InvalidParameterError
	var up *fhlapi.Error
	switch {
	case errors.Is(err, tools.ErrNotImplemented):
		return "Tool '" + tool + "' is not implemented yet in this build"
	case errors.As(err, &ip):
		return ip.Error()
	case errors.As(err, &up):
		return "FHL API request failed - " + up.Error()
	default:
		return err.Error()
	}
}
