package cmds

import (
	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var trackCmd = &cobra.Command{
	Use:     "track [file]",
	Short:   "Track a file with the dotfiles repository",
	Long:    `Track a file with the dotfiles repository`,
	Aliases: []string{"t", "add", "a"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		cfgResult := config.Get()
		if cfgResult.IsErr() {
			color.Red("Failed to get config: %v", cfgResult.UnwrapErr())
			return
		}

		cfg := cfgResult.Value()

		addResult := git.AddFile(cfg, filePath)
		if addResult.IsErr() {
			color.Red("Failed to track file: %v", addResult.Err())
			return
		}

		commitResult := git.CommitStagedFiles(cfg)
		if commitResult.IsErr() {
			color.Red("Failed to commit files: %v", commitResult.Err())
			return
		}

		color.Green("Successfully tracked files")
	},
}

func init() {
	rootCmd.AddCommand(trackCmd)
}
