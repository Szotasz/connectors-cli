package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Szotasz/connectors-cli/internal/api"
	"github.com/Szotasz/connectors-cli/internal/manifest"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Fetch latest tool manifest from connectors.hu",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.Token == "" {
			return fmt.Errorf("CONNECTORS_HU_TOKEN not set. Export your API key first:\n  export CONNECTORS_HU_TOKEN=cnk_your_api_key")
		}

		client := api.New(cfg)

		fmt.Print("Fetching manifest... ")
		m, err := client.FetchManifest()
		if err != nil {
			return fmt.Errorf("fetch failed: %w", err)
		}
		fmt.Printf("OK (%d connectors, %d tools)\n", len(m.Connectors), len(m.Tools))

		if err := manifest.Save(cfg, m); err != nil {
			return fmt.Errorf("save manifest: %w", err)
		}

		if err := manifest.UpdateSkill(m); err != nil {
			fmt.Printf("Warning: could not update Claude skill: %v\n", err)
		} else {
			fmt.Println("Claude Code skill updated.")
		}

		fmt.Println("Done. Run `conn <connector> <command> --help` to see available tools.")
		return nil
	},
}
