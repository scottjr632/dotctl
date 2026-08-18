package cmds

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/srctl/dotctl/internal/git"
)

var addRemoteCmd = &cobra.Command{
	Use:   "add-remote [url]",
	Short: "Add a remote origin if one doesn't exist",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return addRemoteOrigin(cmd, args[0])
	},
}

func init() {
	rootCmd.AddCommand(addRemoteCmd)
}

func addRemoteOrigin(cmd *cobra.Command, url string) error {
	if url == "" {
		return fmt.Errorf("remote URL cannot be empty")
	}

	cfg, err := commandConfig()
	if err != nil {
		return err
	}

	checkCmd := git.GitCmd(cfg, "remote", "get-url", "origin")
	output, err := checkCmd.SilentlyExecute()

	if err == nil {
		return fmt.Errorf("remote origin already exists: %s", strings.TrimSpace(string(output)))
	}

	if dryRun {
		return writePlan(cmd, action("add_remote", fmt.Sprintf("Add remote origin %s to the dotfiles repository", url)))
	}
	addCmd := git.GitCmd(cfg, "remote", "add", "origin", url)
	if output, err := addCmd.SilentlyExecute(); err != nil {
		return fmt.Errorf("failed to add remote origin: %s", string(output))
	}

	if !jsonOutput {
		fmt.Fprintf(cmd.OutOrStdout(), "Successfully added remote origin: %s\n", url)
	}
	return nil
}
