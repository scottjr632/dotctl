package cmds

import (
	"fmt"

	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var listFlag bool

var branchCmd = &cobra.Command{
	Use:   "branch [branch]",
	Short: "Show, list, or switch branches",
	Long:  "Show, list, or switch branches",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 && listFlag {
			return fmt.Errorf("cannot provide a branch and --list together")
		}
		cfg, err := commandConfig()
		if err != nil {
			return err
		}
		if listFlag {
			return git.GitCmd(cfg, "branch", "-a").ExecuteInTerminal()
		}
		if len(args) == 0 {
			return git.GitCmd(cfg, "branch", "--show-current").ExecuteInTerminal()
		}
		if dryRun {
			return writePlan(cmd, fmt.Sprintf("Switch the dotfiles repository to branch %s", args[0]))
		}
		return git.GitCmd(cfg, "switch", args[0]).ExecuteInTerminal()
	},
}

func init() {
	branchCmd.Flags().BoolVarP(&listFlag, "list", "l", false, "list all branches")
	rootCmd.AddCommand(branchCmd)
}
