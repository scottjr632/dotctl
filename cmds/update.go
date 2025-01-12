package cmds

import (
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var updatePatch bool

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
		if updatePatch {
			patchCmd := git.GitCmd(cfg, "add", "-p")
			err := patchCmd.ExecuteInTerminal()
			if err != nil {
				errorPrinter.Println(err)
				return
			}
		} else {
			addCmd := git.GitCmd(cfg, "add", "-u")
			err := addCmd.ExecuteInTerminal()
			if err != nil {
				errorPrinter.Println(err)
				return
			}
		}

		result := git.CommitStagedFiles(cfg)
		if result.IsErr() {
			errorPrinter.Println(result.UnwrapErr())
			return
		}
	},
}

func init() {
	updateCmd.Flags().BoolVarP(&updatePatch, "patch", "p", false, "Use patch to track changes")

	rootCmd.AddCommand(updateCmd)
}
