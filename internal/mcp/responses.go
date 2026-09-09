package mcp

import (
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TextResult creates a successful CallToolResult with a single TextContent item.
func TextResult(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: message,
			},
		},
	}
}

// ErrorResult creates a failed CallToolResult with IsError: true and error message text.
func ErrorResult(format string, args ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf(format, args...),
			},
		},
	}
}

// ScreenshotResult creates a CallToolResult containing both the base64 PNG image and metadata text.
func ScreenshotResult(b64Data []byte, width, height int, deviceID string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.ImageContent{
				Data:     b64Data,
				MIMEType: "image/png",
			},
			&mcp.TextContent{
				Text: fmt.Sprintf("Screenshot captured from device %s (%dx%d)", deviceID, width, height),
			},
		},
	}
}
