package cmds

import (
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage dotfile config",
	Long:  `Manage dotfile config`,
}

var showConfigCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current dotfile config",
	Long:  `Show the current dotfile config`,
	Run: func(cmd *cobra.Command, args []string) {
		config.PrintConfigFile().Must()
	},
}

func init() {
	configCmd.AddCommand(showConfigCmd)
	rootCmd.AddCommand(configCmd)
}
