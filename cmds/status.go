package cmds

import (
	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of the dotfiles repository",
	Long:  `Show the status of the dotfiles repository`,
	Aliases: []string{"st"},
	Run: func(cmd *cobra.Command, args []string) {
		cfgResult := config.Get()
		if cfgResult.IsErr() {
			color.Red("Failed to get config: %v", cfgResult.UnwrapErr())
			return
		}

		cfg := cfgResult.Must()
		statusResult := git.Status(cfg)
		if statusResult.IsErr() {
			color.Red("Failed to get status: %v", statusResult.Err())
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
