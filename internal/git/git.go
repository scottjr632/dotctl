package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

var (
	workTreeOverride string
	nonInteractive   bool
)

type InitRepoOptions struct {
	Path string
}

type StatusInfo struct {
	Branch      string   `json:"branch"`
	Staged      []string `json:"staged"`
	Modified    []string `json:"modified"`
	Ahead       int      `json:"ahead"`
	Behind      int      `json:"behind"`
	HasUpstream bool     `json:"has_upstream"`
}

func SetWorkTree(path string) {
	workTreeOverride = path
}

func SetNonInteractive(value bool) {
	nonInteractive = value
}

func WorkTree() string {
	if workTreeOverride != "" {
		return workTreeOverride
	}
	if path := os.Getenv("DOTCTL_WORK_TREE"); path != "" {
		return path
	}
	if path, err := os.UserHomeDir(); err == nil {
		return path
	}

	return "."
}

func initRepoDefaultOptions(options *InitRepoOptions) {
	if options.Path == "" {
		options.Path = filepath.Join(WorkTree(), ".cfg", ".dotfiles")
	}
}

func GitCmd(cfg config.Config, args ...string) *terminalcmd.Cmd {
	gitDir := "--git-dir=" + filepath.Clean(cfg.DotfilesGitPath)
	cmdArgs := append([]string{gitDir, "--work-tree=" + filepath.Clean(WorkTree())}, args...)
	cmd := terminalcmd.New("git", cmdArgs...)
	if nonInteractive {
		cmd.WithEnv("GIT_TERMINAL_PROMPT=0", "GIT_EDITOR=true")
	}
	return cmd
}

func ConfigureClonedRepo(cfg config.Config) result.Failable {
	commands := [][]string{
		{"config", "--local", "status.showUntrackedFiles", "no"},
		{"config", "--local", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*"},
	}
	for _, args := range commands {
		if err := GitCmd(cfg, args...).ExecuteInTerminal(); err != nil {
			return result.NewFailable(err)
		}
	}

	branch, err := GitCmd(cfg, "symbolic-ref", "--short", "HEAD").SilentlyExecute()
	if err != nil {
		return result.NewFailable(fmt.Errorf("determine cloned repository branch: %w", err))
	}
	branch = strings.TrimSpace(branch)
	remoteRef := "refs/remotes/origin/" + branch
	if err := GitCmd(cfg, "update-ref", remoteRef, "refs/heads/"+branch).ExecuteInTerminal(); err != nil {
		return result.NewFailable(err)
	}
	if err := GitCmd(cfg, "branch", "--set-upstream-to=origin/"+branch, branch).ExecuteInTerminal(); err != nil {
		return result.NewFailable(err)
	}
	return result.NewFailable(nil)
}

func InitBareRepo(options InitRepoOptions) (res result.Failable) {
	initRepoDefaultOptions(&options)
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

func GetStatus(cfg config.Config) result.Result[StatusInfo] {
	branch, err := GitCmd(cfg, "branch", "--show-current").SilentlyExecute()
	if err != nil {
		return result.Err[StatusInfo](err)
	}
	staged, err := GitCmd(cfg, "diff", "--cached", "--name-only", "-z").SilentlyExecute()
	if err != nil {
		return result.Err[StatusInfo](err)
	}
	modified, err := GitCmd(cfg, "diff", "--name-only", "-z").SilentlyExecute()
	if err != nil {
		return result.Err[StatusInfo](err)
	}

	status := StatusInfo{
		Branch:   strings.TrimSpace(branch),
		Staged:   splitNullDelimited(staged),
		Modified: splitNullDelimited(modified),
	}
	if _, err := GitCmd(cfg, "rev-parse", "--verify", "@{upstream}").SilentlyExecute(); err != nil {
		return result.Ok(status)
	}
	counts, err := GitCmd(cfg, "rev-list", "--left-right", "--count", "@{upstream}...HEAD").SilentlyExecute()
	if err != nil {
		return result.Err[StatusInfo](err)
	}

	fields := strings.Fields(counts)
	if len(fields) != 2 {
		return result.Err[StatusInfo](fmt.Errorf("unexpected upstream counts: %q", strings.TrimSpace(counts)))
	}
	behind, behindErr := strconv.Atoi(fields[0])
	ahead, aheadErr := strconv.Atoi(fields[1])
	if behindErr != nil || aheadErr != nil {
		return result.Err[StatusInfo](fmt.Errorf("unexpected upstream counts: %q", strings.TrimSpace(counts)))
	}
	status.Behind = behind
	status.Ahead = ahead
	status.HasUpstream = true
	return result.Ok(status)
}

func ListTrackedFiles(cfg config.Config) result.Result[[]string] {
	output, err := GitCmd(cfg, "ls-files", "-z").SilentlyExecute()
	if err != nil {
		return result.Err[[]string](err)
	}
	return result.Ok(splitNullDelimited(output))
}

func ListHeadFiles(cfg config.Config) result.Result[[]string] {
	output, err := GitCmd(cfg, "ls-tree", "-r", "--name-only", "-z", "HEAD").SilentlyExecute()
	if err != nil {
		return result.Err[[]string](err)
	}
	return result.Ok(splitNullDelimited(output))
}

func IsTracked(cfg config.Config, path string) result.Result[bool] {
	output, err := GitCmd(cfg, "ls-files", "--", path).SilentlyExecute()
	if err != nil {
		return result.Err[bool](err)
	}
	return result.Ok(strings.TrimSpace(output) != "")
}

func splitNullDelimited(output string) []string {
	output = strings.TrimSuffix(output, "\x00")
	if output == "" {
		return []string{}
	}
	return strings.Split(output, "\x00")
}

func AddFile(cfg config.Config, filePath string) result.Failable {
	cmd := GitCmd(cfg, "add", filePath)
	return result.NewFailable(cmd.ExecuteInTerminal())
}

func StageAllFiles(cfg config.Config) result.Failable {
	cmd := GitCmd(cfg, "add", "--all")
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
