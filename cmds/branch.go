package cmds

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/scottjr632/dotctl/internal/utils"
	"github.com/spf13/cobra"
)

var listFlag bool

var branchCmd = &cobra.Command{
	Use:   "branch [branch]",
	Short: "Branch management",
	Long:  "Branch management",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get().Must()

		if len(args) > 0 && listFlag {
			color.Red("Cannot use both branch and list flags")
			os.Exit(1)
			return
		}

		if listFlag {
			cmdErr := git.GitCmd(cfg, "branch", "-a").ExecuteInTerminal()
			if cmdErr != nil {
				fmt.Println(cmdErr)
			}
			return
		}

		if len(args) == 0 {
			cmdErr := git.GitCmd(cfg, "branch", "--show-current").ExecuteInTerminal()
			if cmdErr != nil {
				fmt.Println(cmdErr)
			}
			return
		}

		utils.Invariant(len(args) == 1, "branch only accepts one argument")

		branchName := args[0]
		cmdErr := git.GitCmd(config.Get().Must(), "switch", branchName).ExecuteInTerminal()
		if cmdErr != nil {
			fmt.Println(cmdErr)
			return
		}
	},
}

func init() {
	branchCmd.Flags().BoolVarP(&listFlag, "list", "l", false, "List all branches")
	rootCmd.AddCommand(branchCmd)
}
