package cmds

import (
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tracked files in the dotfiles repository",
	Long:  `List all tracked files in the dotfiles repository`,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		cfgResult := config.Get()
		if cfgResult.IsErr() {
			errorPrinter.Println(cfgResult.UnwrapErr())
			return
		}

		cfg := cfgResult.Must()
		err := git.GitCmd(cfg, "ls-files").ExecuteInTerminal()
		if err != nil {
			errorPrinter.Println(err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
