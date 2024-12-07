package cmds

import (
	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push changes to the remote repository",
	Long:  `Push changes to the remote repository`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgResult := config.Get()
		if cfgResult.IsErr() {
			color.Red("Failed to get config: %v", cfgResult.UnwrapErr())
			return
		}

		cfg := cfgResult.Must()
		pushResult := git.Push(cfg)
		if pushResult.IsErr() {
			color.Red("Failed to push changes: %v", pushResult.Err())
			return
		}

		color.Green("Successfully pushed changes")
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
