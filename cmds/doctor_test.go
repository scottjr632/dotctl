package cmds

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/srctl/dotctl/internal/config"
	"github.com/srctl/dotctl/internal/git"
)

func TestInspectConfigurationUsesOnlyLocalState(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	workTree := filepath.Join(root, "work-tree")
	repository := filepath.Join(root, "repository")
	runnablesDir := filepath.Join(root, "runnables")
	preRunnable := filepath.Join(runnablesDir, "pre.sh")
	for _, dir := range []string{configDir, workTree, runnablesDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if output, err := exec.Command("git", "init", "--bare", repository).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, output)
	}
	if err := os.WriteFile(preRunnable, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(filepath.Join(configDir, "config"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{DotfilesGitPath: repository, DependenciesDir: runnablesDir, PreRunnableFile: preRunnable}
	if err := json.NewEncoder(file).Encode(cfg); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	config.SetDir(configDir)
	git.SetWorkTree(workTree)
	t.Cleanup(func() {
		config.SetDir("")
		git.SetWorkTree("")
	})

	report := inspectConfiguration()
	if !report.Healthy {
		t.Fatalf("expected a healthy report: %+v", report)
	}
	if status := doctorStatus(report, "remote"); status != "warn" {
		t.Fatalf("expected an unconfigured remote warning, got %q: %+v", status, report)
	}
}

func doctorStatus(report doctorReport, name string) string {
	for _, check := range report.Checks {
		if check.Name == name {
			return check.Status
		}
	}
	return ""
}
