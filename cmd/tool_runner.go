package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Szotasz/connectors-cli/internal/api"
)

func makeToolRunner(connectorID, command string, argDefs []api.Arg) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		if cfg.Token == "" {
			return fmt.Errorf("CONNECTORS_HU_TOKEN not set. Export your API key first:\n  export CONNECTORS_HU_TOKEN=cnk_your_api_key")
		}

		args := map[string]interface{}{}

		for _, a := range argDefs {
			switch a.Type {
			case "number":
				v, err := cmd.Flags().GetFloat64(a.Name)
				if err != nil {
					return err
				}
				if cmd.Flags().Changed(a.Name) {
					args[a.Name] = v
				}
			case "boolean":
				v, err := cmd.Flags().GetBool(a.Name)
				if err != nil {
					return err
				}
				if cmd.Flags().Changed(a.Name) {
					args[a.Name] = v
				}
			default:
				v, err := cmd.Flags().GetString(a.Name)
				if err != nil {
					return err
				}
				if v != "" {
					args[a.Name] = v
				}
			}
		}

		client := api.New(cfg)
		resp, err := client.CallTool(connectorID, command, args)
		if err != nil {
			return err
		}

		if resp.Error != nil {
			return fmt.Errorf("API error %d: %s", resp.Error.Code, resp.Error.Message)
		}

		selectField, _ := cmd.Flags().GetString("select")
		csvMode, _ := cmd.Flags().GetBool("csv")

		if selectField != "" || csvMode {
			return formatSelected(resp.Result, selectField, csvMode)
		}

		// MCP `tools/call` result may carry binary content (image, audio) as
		// base64. Dumping that raw to the terminal floods the user's shell —
		// detect MCP `content` blocks and save binaries to disk instead,
		// echoing just the saved path + text/resource_link blocks.
		if handled, err := renderMcpContent(resp.Result, connectorID, command); handled {
			return err
		}

		var pretty bytes.Buffer
		if err := json.Indent(&pretty, resp.Result, "", "  "); err != nil {
			os.Stdout.Write(resp.Result)
		} else {
			pretty.WriteTo(os.Stdout)
		}
		fmt.Println()
		return nil
	}
}

// renderMcpContent inspects the tool result for an MCP `content` array. If
// found, it walks the array: text/resource_link blocks are printed, image/
// audio blocks are saved to ~/Downloads/connectors-<connector>-<ts>.<ext> and
// the saved path is printed instead of the base64 payload.
//
// Returns (true, err) if the result was an MCP content array (handled), or
// (false, nil) to fall through to the default JSON pretty-print path.
func renderMcpContent(raw json.RawMessage, connectorID, command string) (bool, error) {
	var envelope struct {
		Content []map[string]interface{} `json:"content"`
		IsError bool                     `json:"isError"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return false, nil
	}
	if envelope.Content == nil {
		return false, nil
	}

	dlDir, err := defaultDownloadDir()
	if err != nil {
		dlDir = "."
	}
	ts := time.Now().Format("20060102-150405")
	for i, item := range envelope.Content {
		kind, _ := item["type"].(string)
		switch kind {
		case "text":
			if txt, ok := item["text"].(string); ok {
				fmt.Println(txt)
			}
		case "image", "audio":
			data, _ := item["data"].(string)
			mime, _ := item["mimeType"].(string)
			ext := extFromMime(mime)
			fname := fmt.Sprintf("connectors-%s-%s-%d.%s", connectorID, ts, i, ext)
			fpath := filepath.Join(dlDir, fname)
			decoded, decErr := base64.StdEncoding.DecodeString(data)
			if decErr != nil {
				fmt.Printf("⚠️  Failed to decode %s content: %v\n", kind, decErr)
				continue
			}
			if writeErr := os.WriteFile(fpath, decoded, 0o644); writeErr != nil {
				fmt.Printf("⚠️  Failed to save %s to %s: %v\n", kind, fpath, writeErr)
				continue
			}
			label := "🖼"
			if kind == "audio" {
				label = "🎧"
			}
			fmt.Printf("%s  Saved %s (%d bytes) → %s\n", label, mime, len(decoded), fpath)
		case "resource_link":
			uri, _ := item["uri"].(string)
			name, _ := item["name"].(string)
			if name == "" {
				name = "link"
			}
			fmt.Printf("🔗  %s → %s\n", name, uri)
		default:
			b, _ := json.Marshal(item)
			fmt.Printf("ℹ️   %s\n", string(b))
		}
	}
	if envelope.IsError {
		return true, fmt.Errorf("tool returned isError=true")
	}
	return true, nil
}

func defaultDownloadDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dl := filepath.Join(home, "Downloads")
	if info, err := os.Stat(dl); err == nil && info.IsDir() {
		return dl, nil
	}
	return home, nil
}

func extFromMime(mime string) string {
	switch mime {
	case "image/png":
		return "png"
	case "image/jpeg":
		return "jpg"
	case "image/webp":
		return "webp"
	case "image/gif":
		return "gif"
	case "audio/wav":
		return "wav"
	case "audio/mp3", "audio/mpeg":
		return "mp3"
	case "audio/ogg":
		return "ogg"
	case "video/mp4":
		return "mp4"
	}
	if idx := strings.Index(mime, "/"); idx >= 0 && idx+1 < len(mime) {
		return mime[idx+1:]
	}
	return "bin"
}

func formatSelected(raw json.RawMessage, fields string, csvMode bool) error {
	var data interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}

	items := extractItems(data)
	if items == nil {
		os.Stdout.Write(raw)
		fmt.Println()
		return nil
	}

	fieldList := strings.Split(fields, ",")

	if csvMode && len(fieldList) > 0 && fields != "" {
		fmt.Println(strings.Join(fieldList, ","))
	}

	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if fields == "" {
			b, _ := json.Marshal(m)
			fmt.Println(string(b))
			continue
		}
		vals := make([]string, len(fieldList))
		for i, f := range fieldList {
			f = strings.TrimSpace(f)
			if v, ok := m[f]; ok {
				vals[i] = fmt.Sprintf("%v", v)
			}
		}
		if csvMode {
			fmt.Println(strings.Join(vals, ","))
		} else {
			fmt.Println(strings.Join(vals, "\t"))
		}
	}
	return nil
}

func extractItems(data interface{}) []interface{} {
	if arr, ok := data.([]interface{}); ok {
		return arr
	}
	if m, ok := data.(map[string]interface{}); ok {
		for _, v := range m {
			if arr, ok := v.([]interface{}); ok {
				return arr
			}
		}
	}
	return nil
}
