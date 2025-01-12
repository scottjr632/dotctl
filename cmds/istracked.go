package cmds

import (
	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var isTrackedCmd = &cobra.Command{
	Use: "is-tracked [file]",
	Aliases: []string{
		"istracked",
		"is",
		"tracked",
		"it",
	},
	Short: "Check if a file is tracked in the dotfiles repository",
	Long:  `Check if a file is tracked in the dotfiles repository`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		cfgResult := config.Get()
		if cfgResult.IsErr() {
			errorPrinter.Println(cfgResult.UnwrapErr())
			return
		}

		cfg := cfgResult.Must()
		statusCmd := git.GitCmd(cfg, "ls-files", "--error-unmatch", filePath)
		err := statusCmd.ExecuteInTerminal()
		if err != nil {
			color.Red("File '%s' is not tracked", filePath)
		} else {
			color.Green("File '%s' is tracked", filePath)
		}
	},
}

func init() {
	rootCmd.AddCommand(isTrackedCmd)
}
