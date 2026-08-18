package cmds

import (
	"fmt"
	"net/url"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/scottjr632/dotctl/internal/terminalcmd"
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
		remoteURL, err := git.GetRemoteURL(cfg).Unwrap()
		if err != nil {
			return err
		}
		repoURL, err := githubWebURL(remoteURL)
		if err != nil {
			return err
		}
		if dryRun {
			return writePlan(cmd, action("open_url", fmt.Sprintf("Open %s in the default browser", repoURL)))
		}
		executable, arguments, err := browserCommand(runtime.GOOS, repoURL)
		if err != nil {
			return err
		}
		if err := terminalcmd.New(executable, arguments...).ExecuteInTerminal(); err != nil {
			return fmt.Errorf("open repository URL: %w", err)
		}
		if !jsonOutput {
			color.New(color.FgGreen).Fprintln(cmd.OutOrStdout(), "Successfully opened the repository on GitHub")
		}
		return nil
	},
}

func githubWebURL(remote string) (string, error) {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return "", fmt.Errorf("remote URL is not set")
	}
	if strings.HasPrefix(remote, "git@github.com:") {
		return githubHTTPSURL(strings.TrimPrefix(remote, "git@github.com:"))
	}

	parsed, err := url.Parse(remote)
	if err != nil || parsed.Hostname() != "github.com" {
		return "", fmt.Errorf("remote URL is not a GitHub URL")
	}
	return githubHTTPSURL(strings.TrimPrefix(parsed.Path, "/"))
}

func githubHTTPSURL(path string) (string, error) {
	path = strings.TrimSuffix(path, ".git")
	if path == "" || !strings.Contains(path, "/") {
		return "", fmt.Errorf("remote URL does not identify a GitHub repository")
	}
	return "https://github.com/" + path, nil
}

func browserCommand(goos, url string) (string, []string, error) {
	switch goos {
	case "darwin":
		return "open", []string{url}, nil
	case "linux":
		return "xdg-open", []string{url}, nil
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", url}, nil
	default:
		return "", nil, fmt.Errorf("opening a browser is not supported on %s", goos)
	}
}

func init() {
	rootCmd.AddCommand(viewCmd)
}
