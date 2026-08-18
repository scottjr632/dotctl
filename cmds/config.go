package cmds

import (
	"encoding/json"
	"fmt"

	"github.com/scottjr632/dotctl/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage dotfile config",
	Long:  "Manage dotfile config",
}

var showConfigCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current dotfile config",
	Long:  "Show the current dotfile config",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load().Unwrap()
		if err != nil {
			return err
		}
		if wantsJSON(cmd) {
			return writeJSON(cmd, cfg)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Config: %s\n", config.FilePath())
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(cfg)
	},
}

func init() {
	configCmd.AddCommand(showConfigCmd)
	rootCmd.AddCommand(configCmd)
}
