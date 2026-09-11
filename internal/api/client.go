package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Szotasz/connectors-cli/internal/config"
)

type Client struct {
	cfg  *config.Config
	http *http.Client
}

// A 60s ceiling is generous for any single tool call but still bounded —
// without it a hostile or hung upstream server hangs every CLI invocation
// indefinitely (e.g. when Claude Code drives the CLI inside a skill).
const defaultHTTPTimeout = 60 * time.Second

func New(cfg *config.Config) *Client {
	return &Client{
		cfg: cfg,
		http: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
	}
}

func (c *Client) FetchManifest() (*Manifest, error) {
	req, err := http.NewRequest("GET", c.cfg.BaseURL+"/v1/manifest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var m Manifest
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes)).Decode(&m); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}
	return &m, nil
}

// 16 MiB is well above what a real manifest or a single MCP tool response
// would ever need, and small enough that a hostile server can't OOM us.
const maxResponseBytes = 16 * 1024 * 1024

type McpToolCall struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type McpToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type McpResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *McpError       `json:"error,omitempty"`
}

type McpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *Client) CallTool(connectorID, toolName string, args map[string]interface{}) (*McpResponse, error) {
	fullName := connectorID + "_" + toolName

	payload := McpToolCall{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: McpToolCallParams{
			Name:      fullName,
			Arguments: args,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.cfg.BaseURL+"/v1/mcp", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var mcpResp McpResponse
	if err := json.Unmarshal(respBody, &mcpResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &mcpResp, nil
}
