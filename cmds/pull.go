package cmds

import (
	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Fetch and pull the latest data from the repository",
	Long:  `Fetch and pull the latest data from the repository`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgResult := config.Get()
		if cfgResult.IsErr() {
			color.Red("Failed to get config: %v", cfgResult.UnwrapErr())
			return
		}

		cfg := cfgResult.Must()
		err := git.GitCmd(cfg, "pull").ExecuteInTerminal()
		if err != nil {
			color.Red("Failed to pull latest data: %v", err)
			return
		}

		color.Green("Successfully pulled the latest data")
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
