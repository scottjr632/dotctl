package cmds

import (
	"fmt"

	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var checkoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "Check out tracked dotfiles into the work tree",
	Long:  "Check out tracked dotfiles into the work tree",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := commandConfig()
		if err != nil {
			return err
		}
		if dryRun {
			return writePlan(cmd, fmt.Sprintf("Check out tracked files from %s into %s", cfg.DotfilesGitPath, git.WorkTree()))
		}
		if err := git.GitCmd(cfg, "checkout").ExecuteInTerminal(); err != nil {
			return fmt.Errorf("checkout failed; resolve conflicting files before retrying: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkoutCmd)
}
