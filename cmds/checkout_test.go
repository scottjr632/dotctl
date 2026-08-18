package cmds

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCloneAndCheckoutWithBackup(t *testing.T) {
	source := createDotfilesSource(t)
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	workTree := filepath.Join(root, "home")
	repository := filepath.Join(root, "dotfiles.git")
	backupDir := filepath.Join(root, "backup")
	if err := os.MkdirAll(filepath.Join(workTree, ".pi", "agent"), 0o755); err != nil {
		t.Fatal(err)
	}
	settings := filepath.Join(workTree, ".pi", "agent", "settings.json")
	if err := os.WriteFile(settings, []byte("local settings\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	auth := filepath.Join(workTree, ".pi", "agent", "auth.json")
	if err := os.WriteFile(auth, []byte("machine secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	runJSONCommand(t, 0, "--json", "--config-dir", configDir, "--work-tree", workTree, "init", "--clone", source, "--path", repository)
	if got := runGit(t, repository, "rev-parse", "--abbrev-ref", "@{upstream}"); got != "origin/main" {
		t.Fatalf("upstream = %q, want origin/main", got)
	}
	if got := runGit(t, repository, "config", "--get", "remote.origin.fetch"); got != "+refs/heads/*:refs/remotes/origin/*" {
		t.Fatalf("fetch refspec = %q", got)
	}

	plan := runJSONCommand(t, 0, "--json", "--dry-run", "--config-dir", configDir, "--work-tree", workTree, "checkout", "--backup-existing", "--backup-dir", backupDir)
	if plan.Kind != "plan" {
		t.Fatalf("unexpected checkout plan: %+v", plan)
	}
	if _, err := os.Stat(backupDir); !os.IsNotExist(err) {
		t.Fatalf("dry-run created backup directory: %v", err)
	}
	if got := string(mustReadFile(t, settings)); got != "local settings\n" {
		t.Fatalf("dry-run changed settings: %q", got)
	}

	response := runJSONCommand(t, 0, "--json", "--config-dir", configDir, "--work-tree", workTree, "checkout", "--backup-existing", "--backup-dir", backupDir)
	var result checkoutResult
	if err := json.Unmarshal(response.Data, &result); err != nil {
		t.Fatal(err)
	}
	if result.BackupDir != backupDir || !reflect.DeepEqual(result.BackedUp, []string{".pi/agent/settings.json"}) {
		t.Fatalf("unexpected checkout result: %+v", result)
	}
	if got := string(mustReadFile(t, filepath.Join(backupDir, ".pi", "agent", "settings.json"))); got != "local settings\n" {
		t.Fatalf("backup contents = %q", got)
	}
	if got := string(mustReadFile(t, settings)); got != "repository settings\n" {
		t.Fatalf("checked out settings = %q", got)
	}
	if got := string(mustReadFile(t, auth)); got != "machine secret\n" {
		t.Fatalf("untracked auth file changed: %q", got)
	}

	failure := runJSONCommand(t, 1, "--json", "--config-dir", configDir, "--work-tree", workTree, "checkout", "--backup-existing", "--backup-dir", filepath.Join(root, "second-backup"))
	if failure.Error == nil || failure.Error.Code != "INVALID_ARGUMENT" || !strings.Contains(failure.Error.Message, "first checkout") {
		t.Fatalf("unexpected repeated-checkout error: %+v", failure)
	}
}

func TestExistingCheckoutPathsStopsAtSymlink(t *testing.T) {
	workTree := t.TempDir()
	target := t.TempDir()
	if err := os.Symlink(target, filepath.Join(workTree, ".pi")); err != nil {
		t.Fatal(err)
	}

	paths, err := existingCheckoutPaths(workTree, []string{".pi/agent/settings.json", ".pi/agent/theme.json"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(paths, []string{".pi"}) {
		t.Fatalf("existing paths = %#v, want [.pi]", paths)
	}
}

func createDotfilesSource(t *testing.T) string {
	t.Helper()
	source := filepath.Join(t.TempDir(), "source")
	runExternal(t, "git", "init", "--quiet", "--initial-branch=main", source)
	runExternal(t, "git", "-C", source, "config", "user.name", "dotctl test")
	runExternal(t, "git", "-C", source, "config", "user.email", "dotctl@example.com")
	settings := filepath.Join(source, ".pi", "agent", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settings), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settings, []byte("repository settings\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, ".zshrc"), []byte("repository shell\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runExternal(t, "git", "-C", source, "add", ".")
	runExternal(t, "git", "-C", source, "commit", "--quiet", "-m", "initial")
	return source
}

func runGit(t *testing.T, gitDir string, args ...string) string {
	t.Helper()
	return runExternal(t, "git", append([]string{"--git-dir=" + gitDir}, args...)...)
}

func runExternal(t *testing.T, name string, args ...string) string {
	t.Helper()
	command := exec.Command(name, args...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, output)
	}
	return strings.TrimSpace(string(output))
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return contents
}
