package cmds

import (
	"fmt"

	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var commitMessage string

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit staged changes",
	Long:  "Commit staged changes",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := commandConfig()
		if err != nil {
			return err
		}
		if nonInteractive && commitMessage == "" {
			return fmt.Errorf("--message is required in non-interactive mode")
		}
		if dryRun {
			message := commitMessage
			if message == "" {
				message = "<open editor for commit message>"
			}
			return writePlan(cmd, fmt.Sprintf("Commit staged dotfiles with message %q", message))
		}
		if commitMessage != "" {
			return git.CommitWithMessage(cfg, commitMessage).Err()
		}
		return git.CommitStagedFiles(cfg).Err()
	},
}

func init() {
	commitCmd.Flags().StringVarP(&commitMessage, "message", "m", "", "commit message (avoids opening an editor)")
	rootCmd.AddCommand(commitCmd)
}
