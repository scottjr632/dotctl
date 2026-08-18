package cmds

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestProfileApplyLinksSelectedVariant(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	workTree := filepath.Join(root, "home")
	env := newProfileRepo(t, configDir, workTree)

	env.writeTracked(t, ".gitconfig##profile.work", "work config\n")
	env.writeTracked(t, ".gitconfig##profile.personal", "personal config\n")
	target := filepath.Join(workTree, ".gitconfig")

	// No profile is set yet, so neither variant matches and nothing is linked.
	env.run(t, 0, "profile", "apply")
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		t.Fatalf("linked a variant with no matching profile: %v", err)
	}

	env.run(t, 0, "profile", "set", "work")
	plan := env.run(t, 0, "--dry-run", "profile", "apply")
	if plan.Kind != "plan" {
		t.Fatalf("unexpected apply plan: %+v", plan)
	}
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		t.Fatalf("dry run created the link: %v", err)
	}

	env.run(t, 0, "profile", "apply")
	if got := readLink(t, target); got != ".gitconfig##profile.work" {
		t.Fatalf("link target = %q", got)
	}
	if got := string(mustReadFile(t, target)); got != "work config\n" {
		t.Fatalf("linked contents = %q", got)
	}

	// Changing the profile must retarget the link dotctl already owns.
	env.run(t, 0, "profile", "set", "personal")
	env.run(t, 0, "profile", "apply")
	if got := readLink(t, target); got != ".gitconfig##profile.personal" {
		t.Fatalf("link was not retargeted: %q", got)
	}
}

func TestProfileApplyRefusesToClobberExistingFile(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	workTree := filepath.Join(root, "home")
	env := newProfileRepo(t, configDir, workTree)

	env.writeTracked(t, ".zshrc##profile.work", "variant shell\n")
	env.run(t, 0, "profile", "set", "work")
	target := filepath.Join(workTree, ".zshrc")
	if err := os.WriteFile(target, []byte("existing shell\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	failure := env.run(t, 1, "profile", "apply")
	if failure.Error == nil || failure.Error.Code != "PROFILE_CONFLICT" {
		t.Fatalf("unexpected conflict response: %+v", failure)
	}
	if got := string(mustReadFile(t, target)); got != "existing shell\n" {
		t.Fatalf("conflicting file was modified: %q", got)
	}

	env.run(t, 0, "profile", "apply", "--force")
	if got := readLink(t, target); got != ".zshrc##profile.work" {
		t.Fatalf("link target = %q", got)
	}
}

// A tracked plain path cannot be replaced by a link: doing so would leave the
// repository permanently dirty, so --force must not override it either.
func TestProfileApplyRefusesToReplaceTrackedTarget(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	workTree := filepath.Join(root, "home")
	env := newProfileRepo(t, configDir, workTree)

	env.writeTracked(t, ".vimrc", "plain vimrc\n")
	env.writeTracked(t, ".vimrc##os."+runtime.GOOS, "variant vimrc\n")

	failure := env.run(t, 1, "profile", "apply", "--force")
	if failure.Error == nil || failure.Error.Code != "PROFILE_CONFLICT" {
		t.Fatalf("unexpected conflict response: %+v", failure)
	}
	if !strings.Contains(failure.Error.Message, "untrack") {
		t.Fatalf("conflict message does not explain the fix: %q", failure.Error.Message)
	}
}

// Checkout links variants itself so that setting up a new machine is a single
// command, which is the flow the whole feature exists for.
func TestCheckoutLinksSelectedVariants(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	workTree := filepath.Join(root, "home")
	env := newProfileRepo(t, configDir, workTree)

	env.writeTracked(t, ".gitconfig##profile.work", "work config\n")
	env.run(t, 0, "profile", "set", "work")
	env.run(t, 0, "checkout")
	if got := readLink(t, filepath.Join(workTree, ".gitconfig")); got != ".gitconfig##profile.work" {
		t.Fatalf("checkout did not link the variant: %q", got)
	}

	if err := os.Remove(filepath.Join(workTree, ".gitconfig")); err != nil {
		t.Fatal(err)
	}
	env.run(t, 0, "checkout", "--skip-profile")
	if _, err := os.Lstat(filepath.Join(workTree, ".gitconfig")); !os.IsNotExist(err) {
		t.Fatalf("--skip-profile still linked the variant: %v", err)
	}
}

func TestProfileListReportsCandidates(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	workTree := filepath.Join(root, "home")
	env := newProfileRepo(t, configDir, workTree)

	env.writeTracked(t, ".gitconfig##profile.work", "work config\n")
	env.writeTracked(t, ".gitconfig##profile.personal", "personal config\n")
	env.run(t, 0, "profile", "set", "work")

	response := env.run(t, 0, "profile", "list")
	var resolution struct {
		Targets []struct {
			Path       string `json:"path"`
			Source     string `json:"source"`
			Candidates []struct {
				Source   string `json:"source"`
				Selected bool   `json:"selected"`
			} `json:"candidates"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(response.Data, &resolution); err != nil {
		t.Fatal(err)
	}
	if len(resolution.Targets) != 1 || resolution.Targets[0].Path != ".gitconfig" {
		t.Fatalf("unexpected targets: %+v", resolution.Targets)
	}
	if resolution.Targets[0].Source != ".gitconfig##profile.work" {
		t.Fatalf("selected %q", resolution.Targets[0].Source)
	}
	if len(resolution.Targets[0].Candidates) != 2 {
		t.Fatalf("unexpected candidates: %+v", resolution.Targets[0].Candidates)
	}
}

type profileRepo struct {
	configDir string
	workTree  string
}

func newProfileRepo(t *testing.T, configDir, workTree string) profileRepo {
	t.Helper()
	if err := os.MkdirAll(workTree, 0o755); err != nil {
		t.Fatal(err)
	}
	env := profileRepo{configDir: configDir, workTree: workTree}
	env.run(t, 0, "init")
	env.run(t, 0, "git", "config", "user.name", "dotctl test")
	env.run(t, 0, "git", "config", "user.email", "dotctl@example.com")
	env.run(t, 0, "git", "config", "commit.gpgsign", "false")
	return env
}

func (r profileRepo) run(t *testing.T, expectedExit int, args ...string) contractResponse {
	t.Helper()
	global := []string{"--json", "--config-dir", r.configDir, "--work-tree", r.workTree}
	return runJSONCommand(t, expectedExit, append(global, args...)...)
}

func (r profileRepo) writeTracked(t *testing.T, name, contents string) {
	t.Helper()
	path := filepath.Join(r.workTree, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	r.run(t, 0, "track", name, "--message", "Track "+name)
}

func readLink(t *testing.T, path string) string {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s is not a symlink", path)
	}
	target, err := os.Readlink(path)
	if err != nil {
		t.Fatal(err)
	}
	return target
}
