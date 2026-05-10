package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Szotasz/connectors-cli/internal/config"
	"github.com/Szotasz/connectors-cli/internal/manifest"
)

var (
	cfg     *config.Config
	Version = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "connectors",
	Short: "connectors.hu CLI -- Hungarian business API gateway",
	Long:  "Query and manage Hungarian business APIs (Billingo, NAV, MiniCRM) through connectors.hu.",
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	cfg = config.Load()
}

func Execute() {
	loadDynamicCommands()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func loadDynamicCommands() {
	cfg = config.Load()
	m, err := manifest.Load(cfg)
	if err != nil {
		return
	}

	connectorCmds := map[string]*cobra.Command{}

	for _, c := range m.Connectors {
		cc := &cobra.Command{
			Use:   c.ID,
			Short: c.Name + " -- " + c.Description,
		}
		connectorCmds[c.ID] = cc
		rootCmd.AddCommand(cc)
	}

	for _, tool := range m.Tools {
		parent, ok := connectorCmds[tool.Connector]
		if !ok {
			continue
		}

		t := tool
		toolCmd := &cobra.Command{
			Use:   dashToUnderscore(t.Command),
			Short: t.Description,
			RunE:  makeToolRunner(t.Connector, t.Command, t.Args),
		}

		for _, arg := range t.Args {
			switch arg.Type {
			case "number":
				toolCmd.Flags().Float64(arg.Name, 0, arg.Description)
			case "boolean":
				toolCmd.Flags().Bool(arg.Name, false, arg.Description)
			default:
				toolCmd.Flags().String(arg.Name, "", arg.Description)
			}
			if arg.Required {
				_ = toolCmd.MarkFlagRequired(arg.Name)
			}
		}

		toolCmd.Flags().String("select", "", "Comma-separated fields to extract from response")
		toolCmd.Flags().Bool("csv", false, "Output selected fields as CSV")

		parent.AddCommand(toolCmd)
	}
}

func dashToUnderscore(s string) string {
	return s
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("connectors", Version)
	},
}
