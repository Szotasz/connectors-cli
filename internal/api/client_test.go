package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Szotasz/connectors-cli/internal/config"
)

// jsonOfSize returns a valid JSON object padded with trailing whitespace to
// exactly n bytes, so the size, not the syntax, decides the outcome.
func jsonOfSize(prefix string, n int) string {
	return prefix + strings.Repeat(" ", n-len(prefix))
}

func serve(t *testing.T, body string) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return New(&config.Config{BaseURL: srv.URL, Token: "test-token"})
}

func TestResponseSizeLimit(t *testing.T) {
	manifest := `{"connectors":[],"tools":[]}`
	mcp := `{"jsonrpc":"2.0","id":1,"result":{}}`

	t.Run("manifest at exactly the limit is accepted", func(t *testing.T) {
		if _, err := serve(t, jsonOfSize(manifest, maxResponseBytes)).FetchManifest(); err != nil {
			t.Fatalf("FetchManifest at the limit: %v", err)
		}
	})
	t.Run("manifest one byte over the limit fails loudly", func(t *testing.T) {
		_, err := serve(t, jsonOfSize(manifest, maxResponseBytes+1)).FetchManifest()
		if !errors.Is(err, ErrResponseTooLarge) || !strings.Contains(err.Error(), "response exceeds 16 MiB") {
			t.Fatalf("FetchManifest over the limit: got %v, want ErrResponseTooLarge", err)
		}
	})
	t.Run("tool call at exactly the limit is accepted", func(t *testing.T) {
		if _, err := serve(t, jsonOfSize(mcp, maxResponseBytes)).CallTool("billingo", "list_documents", nil); err != nil {
			t.Fatalf("CallTool at the limit: %v", err)
		}
	})
	t.Run("tool call one byte over the limit fails loudly", func(t *testing.T) {
		_, err := serve(t, jsonOfSize(mcp, maxResponseBytes+1)).CallTool("billingo", "list_documents", nil)
		if !errors.Is(err, ErrResponseTooLarge) || !strings.Contains(err.Error(), "response exceeds 16 MiB") {
			t.Fatalf("CallTool over the limit: got %v, want ErrResponseTooLarge", err)
		}
	})
}

func TestTimeoutFollowsTheServerWallClock(t *testing.T) {
	c := New(&config.Config{BaseURL: "https://example.invalid"})
	if c.http.Timeout.Seconds() != 300 {
		t.Fatalf("HTTP timeout = %v, want 300s (the Edge Function wall clock)", c.http.Timeout)
	}
}
