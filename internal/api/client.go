package api

import (
	"bytes"
	"encoding/json"
	"errors"
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

// The ceiling follows the server's wall clock, not a guess at a typical call.
// A tool call runs inside a Supabase Edge Function, which is cut at 300s, and
// long-running connectors are written against that budget: the fal.ai client
// polls for up to 270s, and billingo download_document_export alone runs 30
// polls with a 2s sleep (60s) before any real work. A 60s client timeout cut
// those calls off while the server was still working on them. 300s keeps the
// call bounded (a hung upstream cannot hang the CLI forever, e.g. when Claude
// Code drives it inside a skill) without being shorter than the server.
const defaultHTTPTimeout = 300 * time.Second

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

	body, err := readLimited(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("manifest: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var m Manifest
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}
	return &m, nil
}

// 16 MiB is well above what a real manifest or a single MCP tool response
// would ever need, and small enough that a hostile server can't OOM us.
const maxResponseBytes = 16 * 1024 * 1024

// ErrResponseTooLarge is returned instead of a silently truncated body.
var ErrResponseTooLarge = errors.New("response exceeds 16 MiB")

// readLimited reads at most maxResponseBytes. It reads one byte past the
// limit, so a body that is longer fails loudly instead of being cut at the
// limit and handed on as if it were complete.
func readLimited(r io.Reader) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(r, maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if len(body) > maxResponseBytes {
		return nil, ErrResponseTooLarge
	}
	return body, nil
}

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

	respBody, err := readLimited(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("tools/call %s: %w", fullName, err)
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
