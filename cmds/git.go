package cmds

import (
	"fmt"

	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
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
			return writePlan(cmd, fmt.Sprintf("Run Git against the dotfiles repository with arguments %q", args))
		}
		return git.GitCmd(cfg, args...).ExecuteInTerminal()
	},
}

func init() {
	rootCmd.AddCommand(gitCmd)
}
