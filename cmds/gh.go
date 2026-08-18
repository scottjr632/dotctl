package cmds

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:     "view",
	Short:   "Open the git repository on GitHub",
	Long:    `Open the git repository on GitHub in the default web browser`,
	Aliases: []string{"gh", "open"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := commandConfig()
		if err != nil {
			return err
		}
		repoURL, err := git.GetRemoteURL(cfg).Unwrap()
		if err != nil {
			return err
		}
		if repoURL == "" {
			return fmt.Errorf("remote URL is not set")
		}
		if !strings.Contains(repoURL, "github.com") {
			return fmt.Errorf("remote URL is not a GitHub URL")
		}
		if dryRun {
			return writePlan(cmd, fmt.Sprintf("Open %s in the default browser", repoURL))
		}
		if err := exec.Command("open", repoURL).Start(); err != nil {
			return fmt.Errorf("open repository URL: %w", err)
		}
		color.New(color.FgGreen).Fprintln(cmd.OutOrStdout(), "Successfully opened the repository on GitHub")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
}
