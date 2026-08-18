package git

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/terminalcmd"
)

func TestRepositoryInspection(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	remote := filepath.Join(root, "remote")
	workTree := filepath.Join(root, "work-tree")
	if err := os.Mkdir(workTree, 0o755); err != nil {
		t.Fatal(err)
	}
	runSystemGit(t, "init", "--bare", "--initial-branch=main", repo)
	runSystemGit(t, "init", "--bare", "--initial-branch=main", remote)

	SetWorkTree(workTree)
	t.Cleanup(func() { SetWorkTree("") })
	cfg := config.Config{DotfilesGitPath: repo}
	runDotfilesGit(t, cfg, "config", "user.name", "dotctl test")
	runDotfilesGit(t, cfg, "config", "user.email", "dotctl@example.com")
	runDotfilesGit(t, cfg, "config", "commit.gpgsign", "false")

	file := filepath.Join(workTree, "example.txt")
	if err := os.WriteFile(file, []byte("first\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runDotfilesGit(t, cfg, "add", "example.txt")
	runDotfilesGit(t, cfg, "commit", "-m", "initial")
	runDotfilesGit(t, cfg, "remote", "add", "origin", remote)
	runDotfilesGit(t, cfg, "push", "--set-upstream", "origin", "main")

	files, err := ListTrackedFiles(cfg).Unwrap()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(files, []string{"example.txt"}) {
		t.Fatalf("unexpected tracked files: %#v", files)
	}
	tracked, err := IsTracked(cfg, "example.txt").Unwrap()
	if err != nil || !tracked {
		t.Fatalf("expected tracked file, tracked=%v err=%v", tracked, err)
	}

	if err := os.WriteFile(file, []byte("second\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	status, err := GetStatus(cfg).Unwrap()
	if err != nil {
		t.Fatal(err)
	}
	if status.Branch != "main" || !reflect.DeepEqual(status.Modified, []string{"example.txt"}) {
		t.Fatalf("unexpected status: %+v", status)
	}
	if !status.HasUpstream || status.Ahead != 0 || status.Behind != 0 {
		t.Fatalf("unexpected upstream status: %+v", status)
	}

	runDotfilesGit(t, cfg, "add", "example.txt")
	runDotfilesGit(t, cfg, "commit", "-m", "local change")
	status, err = GetStatus(cfg).Unwrap()
	if err != nil {
		t.Fatal(err)
	}
	if status.Ahead != 1 || status.Behind != 0 {
		t.Fatalf("unexpected ahead/behind counts: %+v", status)
	}
}

func runSystemGit(t *testing.T, args ...string) {
	t.Helper()
	output, err := terminalcmd.New("git", args...).SilentlyExecute()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}

func runDotfilesGit(t *testing.T, cfg config.Config, args ...string) {
	t.Helper()
	output, err := GitCmd(cfg, args...).SilentlyExecute()
	if err != nil {
		t.Fatalf("dotfiles git %v: %v\n%s", args, err, output)
	}
}
