package cmds

import (
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update and commit all tracked files that have changes",
	Long:  `Update and commit all tracked files that have changes in the dotfiles repository`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgResult := config.Get()
		if cfgResult.IsErr() {
			errorPrinter.Println(cfgResult.UnwrapErr())
			return
		}

		cfg := cfgResult.Must()
		addCmd := git.GitCmd(cfg, "add", "-u")
		err := addCmd.ExecuteInTerminal()
		if err != nil {
			errorPrinter.Println(err)
			return
		}

		result := git.CommitStagedFiles(cfg)
		if result.IsErr() {
			errorPrinter.Println(result.UnwrapErr())
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
