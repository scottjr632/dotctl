package git

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/result"
	"github.com/scottjr632/dotctl/internal/terminalcmd"
	"github.com/scottjr632/dotctl/internal/utils"
)

type NonEmptyDirError struct {
	error
}

func IsNonEmptyDirError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(NonEmptyDirError)
	return ok
}

type InitRepoOptions struct {
	Path string
}

func initRepoDefaultOptions(options *InitRepoOptions) {
	if options.Path == "" {
		homePath, err := os.UserHomeDir()
		if err != nil {
			slog.Error("Failed to get home directory", "error", err)
			options.Path = "./.cfg/.dotfiles" // fallback to relative path
		} else {
			options.Path = fmt.Sprintf("%s/.cfg/.dotfiles", homePath)
		}
	}
}

func GitCmd(cfg config.Config, args ...string) *terminalcmd.Cmd {
	gitDir := fmt.Sprintf("--git-dir=%s", cfg.DotfilesGitPath)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Error("Failed to get home directory", "error", err)
		return nil
	}
	cmdArgs := append([]string{gitDir, "--work-tree=" + homeDir}, args...)
	return terminalcmd.New("git", cmdArgs...)
}

func InitBareRepo(options InitRepoOptions) (res result.Failable) {
	initRepoDefaultOptions(&options)
	slog.Info("Initializing dotfile config", "path", options.Path)

	err := os.MkdirAll(options.Path, 0755)
	if err != nil {
		return result.NewFailable(err)
	}

	if isDirEmpty, err := utils.IsDirectoryEmpty(options.Path); err != nil {
		return result.NewFailable(err)
	} else if !isDirEmpty {
		return result.NewFailable(NonEmptyDirError{errors.New("directory is not empty")})
	}

	err = terminalcmd.New("git", "init", "--bare", options.Path).ExecuteInTerminal()
	if err != nil {
		return result.NewFailable(err)
	}

	err = GitCmd(config.Config{
		DotfilesGitPath: options.Path,
	}, "config", "--local", "status.showUntrackedFiles", "no").ExecuteInTerminal()
	return result.NewFailable(err)
}

func Status(cfg config.Config) result.Failable {
	cmd := GitCmd(cfg, "status")
	return result.NewFailable(cmd.ExecuteInTerminal())
}

func AddFile(cfg config.Config, filePath string) result.Failable {
	cmd := GitCmd(cfg, "add", filePath)
	return result.NewFailable(cmd.ExecuteInTerminal())
}

func Push(cfg config.Config) result.Failable {
	cmd := GitCmd(cfg, "push")
	return result.NewFailable(cmd.ExecuteInTerminal())
}

func CommitFile(cfg config.Config, filePath string) result.Failable {
	// Check if the file is already tracked by git
	statusCmd := GitCmd(cfg, "ls-files", "--error-unmatch", filePath)
	err := statusCmd.ExecuteInTerminal()

	var commitMessage string
	if err != nil {
		// File is not tracked, so we add it
		addCmd := GitCmd(cfg, "add", filePath)
		if err := addCmd.ExecuteInTerminal(); err != nil {
			return result.NewFailable(err)
		}
		commitMessage = "Add " + filePath
	} else {
		// File is already tracked, so we update it
		commitMessage = "Update " + filePath
	}

	// Commit the changes
	commitCmd := GitCmd(cfg, "commit", "-m", commitMessage)
	return result.NewFailable(commitCmd.ExecuteInTerminal())
}

func CommitWithMessage(cfg config.Config, message string) result.Failable {
	cmd := GitCmd(cfg, "commit", "-m", message)
	return result.NewFailable(cmd.ExecuteInTerminal())
}

func CommitStagedFiles(cfg config.Config) result.Failable {
	cmd := GitCmd(cfg, "commit")
	return result.NewFailable(cmd.ExecuteInTerminal())
}

func GetStagedFiles(cfg config.Config) result.Result[[]string] {
	cmd := GitCmd(cfg, "diff", "--cached", "--name-only")
	out, err := cmd.SilentlyExecute()
	if err != nil {
		return result.Err[[]string](err)
	}
	return result.Ok(strings.Split(out, "\n"))
}

func ResetAllStagedFiles(cfg config.Config) result.Failable {
	cmd := GitCmd(cfg, "reset")
	return result.NewFailable(cmd.ExecuteInTerminal())
}

func GetRemoteURL(cfg config.Config) result.Result[string] {
	cmd := GitCmd(cfg, "remote", "get-url", "origin")
	output, err := cmd.SilentlyExecute()
	if err != nil {
		return result.Err[string](err)
	}
	return result.Ok(strings.TrimSpace(string(output)))
}
