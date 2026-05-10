package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
