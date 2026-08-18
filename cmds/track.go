package cmds

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var trackMessage string

var trackCmd = &cobra.Command{
	Use:     "track [file]",
	Short:   "Track and commit a file in the dotfiles repository",
	Long:    "Track and commit a file in the dotfiles repository",
	Aliases: []string{"t", "add", "a"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := commandConfig()
		if err != nil {
			return err
		}
		if nonInteractive && trackMessage == "" {
			return fmt.Errorf("--message is required in non-interactive mode")
		}
		if dryRun {
			message := trackMessage
			if message == "" {
				message = "<open editor for commit message>"
			}
			return writePlan(cmd,
				fmt.Sprintf("Stage %s in the dotfiles repository", args[0]),
				fmt.Sprintf("Commit staged dotfiles with message %q", message),
			)
		}
		if result := git.AddFile(cfg, args[0]); result.IsErr() {
			return result.Err()
		}
		var resultErr error
		if trackMessage == "" {
			resultErr = git.CommitStagedFiles(cfg).Err()
		} else {
			resultErr = git.CommitWithMessage(cfg, trackMessage).Err()
		}
		if resultErr != nil {
			return resultErr
		}
		color.New(color.FgGreen).Fprintln(cmd.OutOrStdout(), "Successfully tracked file")
		return nil
	},
}

func init() {
	trackCmd.Flags().StringVarP(&trackMessage, "message", "m", "", "commit message (avoids opening an editor)")
	rootCmd.AddCommand(trackCmd)
}
