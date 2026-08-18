package cmds

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/srctl/dotctl/internal/git"
)

var gitCmd = &cobra.Command{
	Use:                "git [git-args]",
	Short:              "Run Git against the dotfiles repository",
	Long:               "Run Git against the dotfiles repository",
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := commandConfig()
		if err != nil {
			return err
		}
		if dryRun {
			return writePlan(cmd, action("git", fmt.Sprintf("Run Git against the dotfiles repository with arguments %q", args)))
		}
		return git.GitCmd(cfg, args...).ExecuteInTerminal()
	},
}

func init() {
	rootCmd.AddCommand(gitCmd)
}
