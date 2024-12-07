package cmds

import (
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var ghCmd = &cobra.Command{
	Use:   "gh",
	Short: "Open the git repository on GitHub",
	Long:  `Open the git repository on GitHub in the default web browser`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgResult := config.Get()
		if cfgResult.IsErr() {
			color.Red("Failed to get config: %v", cfgResult.UnwrapErr())
			return
		}

		cfg := cfgResult.Must()
		remoteURLResult := git.GetRemoteURL(cfg)
		if remoteURLResult.IsErr() {
			color.Red("Failed to get remote URL: %v", remoteURLResult.UnwrapErr())
			return
		}

		repoURL := remoteURLResult.Must()
		if repoURL == "" {
			color.Red("Remote URL is not set")
			return
		}

		// Ensure the URL is a GitHub URL
		if !strings.Contains(repoURL, "github.com") {
			color.Red("The remote URL is not a GitHub URL")
			return
		}

		err := exec.Command("open", repoURL).Start()
		if err != nil {
			color.Red("Failed to open the repository URL: %v", err)
			return
		}

		color.Green("Successfully opened the repository on GitHub")
	},
}

func init() {
	rootCmd.AddCommand(ghCmd)
}
