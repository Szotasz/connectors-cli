package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Szotasz/conn-cli/internal/api"
)

func UpdateSkill(m *api.Manifest) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	skillDir := filepath.Join(home, ".claude", "skills", "conn-hu")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString("name: conn-hu\n")
	sb.WriteString("description: CLI for connectors.hu -- Hungarian business API gateway (Billingo, NAV, MiniCRM). Use when user asks to query invoices, partners, tax data, or any connectors.hu operation.\n")
	sb.WriteString("---\n\n")
	sb.WriteString("# conn -- connectors.hu CLI\n\n")
	sb.WriteString("Run `conn <connector> <command> [flags]` to call connectors.hu APIs.\n\n")

	for _, c := range m.Connectors {
		sb.WriteString(fmt.Sprintf("## %s\n\n", c.Name))

		for _, t := range m.Tools {
			if t.Connector != c.ID {
				continue
			}
			sb.WriteString(fmt.Sprintf("### conn %s %s\n\n", c.ID, t.Command))
			sb.WriteString(t.Description + "\n\n")

			if len(t.Args) > 0 {
				sb.WriteString("Flags:\n")
				for _, a := range t.Args {
					req := ""
					if a.Required {
						req = " (required)"
					}
					desc := a.Description
					if desc == "" {
						desc = a.Type
					}
					if len(a.Enum) > 0 {
						desc += " [" + strings.Join(a.Enum, "|") + "]"
					}
					sb.WriteString(fmt.Sprintf("  --%s  %s%s\n", a.Name, desc, req))
				}
				sb.WriteString("\n")
			}
		}
	}

	sb.WriteString("## Usage\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("export CONN_HU_TOKEN=cnk_your_api_key\n")
	sb.WriteString("conn sync              # fetch latest tool manifest\n")
	sb.WriteString("conn billingo list-documents --per_page 5\n")
	sb.WriteString("conn nav query-taxpayer --taxNumber 12345678\n")
	sb.WriteString("```\n")

	return os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(sb.String()), 0644)
}
