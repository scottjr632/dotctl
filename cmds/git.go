package cmds

import (
	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var gitCmd = &cobra.Command{
	Use:   "git [git-args]",
	Short: "Pass through git commands to git",
	Long:  `Pass through git commands to git`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			color.Red("No git command provided")
			return
		}

		cfgResult := config.Get()
		if cfgResult.IsErr() {
			color.Red("Failed to get config: %v", cfgResult.UnwrapErr())
			return
		}

		cfg := cfgResult.Must()
		gitCmd := git.GitCmd(cfg, args...)
		if gitCmd == nil {
			color.Red("Failed to create git command")
			return
		}

		err := gitCmd.ExecuteInTerminal()
		if err != nil {
			color.Red("Failed to execute git command: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(gitCmd)
}
