package cmds

import (
	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show git log",
	Long:  `Show git log`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgResult := config.Get()
		if cfgResult.IsErr() {
			color.Red("Failed to get config: %v", cfgResult.UnwrapErr())
			return
		}

		cfg := cfgResult.Must()
		logResult := git.GitCmd(cfg, "log", "--name-only").ExecuteInTerminal()
		if logResult != nil {
			color.Red("Failed to get git log: %v", logResult)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}
