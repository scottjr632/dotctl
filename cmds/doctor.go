package cmds

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var errDoctorUnhealthy = errors.New("dotctl configuration is not healthy")

type doctorCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type doctorReport struct {
	Healthy bool          `json:"healthy"`
	Checks  []doctorCheck `json:"checks"`
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check dotctl configuration without changing it",
	Long:  "Check dotctl configuration, paths, and local Git repository without changing them or accessing the network",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		report := inspectConfiguration()
		if wantsJSON(cmd) {
			if err := writeJSONResponse(cmd, report.Healthy, report); err != nil {
				return err
			}
		} else {
			for _, check := range report.Checks {
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s: %s\n", check.Status, check.Name, check.Message)
			}
		}
		if !report.Healthy {
			return errDoctorUnhealthy
		}
		return nil
	},
}

func inspectConfiguration() doctorReport {
	report := doctorReport{Healthy: true, Checks: []doctorCheck{}}
	addCheck := func(name, status, message string) {
		report.Checks = append(report.Checks, doctorCheck{Name: name, Status: status, Message: message})
		if status == "fail" {
			report.Healthy = false
		}
	}

	if path, err := exec.LookPath("git"); err != nil {
		addCheck("git", "fail", "Git is not available on PATH")
	} else {
		addCheck("git", "pass", path)
	}
	checkDirectory("work_tree", git.WorkTree(), addCheck)

	cfg, err := config.Load().Unwrap()
	if err != nil {
		addCheck("config", "fail", fmt.Sprintf("%s: %v", config.FilePath(), err))
		return report
	}
	addCheck("config", "pass", config.FilePath())

	if cfg.DotfilesGitPath == "" {
		addCheck("git_repository", "fail", "git_repo_path is empty")
	} else if info, err := os.Stat(cfg.DotfilesGitPath); err != nil {
		addCheck("git_repository", "fail", fmt.Sprintf("%s: %v", cfg.DotfilesGitPath, err))
	} else if !info.IsDir() {
		addCheck("git_repository", "fail", fmt.Sprintf("%s is not a directory", cfg.DotfilesGitPath))
	} else if _, err := git.GitCmd(cfg, "rev-parse", "--git-dir").SilentlyExecute(); err != nil {
		addCheck("git_repository", "fail", err.Error())
	} else {
		addCheck("git_repository", "pass", cfg.DotfilesGitPath)
	}

	if cfg.DependenciesDir == "" {
		addCheck("runnables", "fail", "dependencies_dir is empty")
	} else {
		checkDirectory("runnables", cfg.DependenciesDir, addCheck)
	}

	if cfg.PreRunnableFile == "" {
		addCheck("pre_runnable", "fail", "pre_runnable_file is empty")
	} else if info, err := os.Stat(cfg.PreRunnableFile); err != nil {
		addCheck("pre_runnable", "fail", fmt.Sprintf("%s: %v", cfg.PreRunnableFile, err))
	} else if !info.Mode().IsRegular() {
		addCheck("pre_runnable", "fail", fmt.Sprintf("%s is not a regular file", cfg.PreRunnableFile))
	} else if info.Mode().Perm()&0o111 == 0 {
		addCheck("pre_runnable", "fail", fmt.Sprintf("%s is not executable", cfg.PreRunnableFile))
	} else {
		addCheck("pre_runnable", "pass", cfg.PreRunnableFile)
	}

	if report.Healthy {
		if remote, err := git.GetRemoteURL(cfg).Unwrap(); err != nil {
			addCheck("remote", "warn", "origin is not configured")
		} else {
			addCheck("remote", "pass", remote)
		}
	}
	return report
}

func checkDirectory(name, path string, addCheck func(string, string, string)) {
	info, err := os.Stat(path)
	if err != nil {
		addCheck(name, "fail", fmt.Sprintf("%s: %v", path, err))
		return
	}
	if !info.IsDir() {
		addCheck(name, "fail", fmt.Sprintf("%s is not a directory", path))
		return
	}
	addCheck(name, "pass", path)
}

func init() {
	addJSONFlag(doctorCmd)
	rootCmd.AddCommand(doctorCmd)
}
