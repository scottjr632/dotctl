package cmds

import (
	"fmt"
	"strings"

	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var addRemoteCmd = &cobra.Command{
	Use:   "add-remote [url]",
	Short: "Add a remote origin if one doesn't exist",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return addRemoteOrigin(args[0])
	},
}

func init() {
	rootCmd.AddCommand(addRemoteCmd)
}

func addRemoteOrigin(url string) error {
	if url == "" {
		return fmt.Errorf("remote URL cannot be empty")
	}

	cfg := config.Get().Must()

	checkCmd := git.GitCmd(cfg, "remote", "get-url", "origin")
	output, err := checkCmd.SilentlyExecute()

	if err == nil {
		return fmt.Errorf("remote origin already exists: %s", strings.TrimSpace(string(output)))
	}

	addCmd := git.GitCmd(cfg, "remote", "add", "origin", url)
	if output, err := addCmd.SilentlyExecute(); err != nil {
		return fmt.Errorf("failed to add remote origin: %s", string(output))
	}

	fmt.Printf("Successfully added remote origin: %s\n", url)
	return nil
}
