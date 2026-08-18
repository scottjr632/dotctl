package cmds

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/srctl/dotctl/internal/git"
)

var (
	updatePatch   bool
	updateMessage string
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update and commit changed tracked files",
	Long:  "Update and commit changed tracked files in the dotfiles repository",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := commandConfig()
		if err != nil {
			return err
		}
		if nonInteractive && updatePatch {
			return fmt.Errorf("--patch is unavailable in non-interactive mode")
		}
		if nonInteractive && updateMessage == "" {
			return fmt.Errorf("--message is required in non-interactive mode")
		}
		if dryRun {
			stageAction := "Stage changes to all tracked dotfiles"
			if updatePatch {
				stageAction = "Interactively select tracked dotfile changes to stage"
			}
			message := updateMessage
			if message == "" {
				message = "<open editor for commit message>"
			}
			return writePlan(cmd,
				action("stage", stageAction),
				action("commit", fmt.Sprintf("Commit staged dotfiles with message %q", message)),
			)
		}
		addArgs := []string{"add", "-u"}
		if updatePatch {
			addArgs = []string{"add", "-p"}
		}
		if err := git.GitCmd(cfg, addArgs...).ExecuteInTerminal(); err != nil {
			return fmt.Errorf("stage updates: %w", err)
		}
		if updateMessage != "" {
			return git.CommitWithMessage(cfg, updateMessage).Err()
		}
		return git.CommitStagedFiles(cfg).Err()
	},
}

func init() {
	updateCmd.Flags().BoolVarP(&updatePatch, "patch", "p", false, "interactively select changes to stage")
	updateCmd.Flags().StringVarP(&updateMessage, "message", "m", "", "commit message (avoids opening an editor)")
	rootCmd.AddCommand(updateCmd)
}
