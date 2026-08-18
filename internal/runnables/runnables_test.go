package runnables

import (
	"path/filepath"
	"testing"

	"github.com/srctl/dotctl/internal/config"
)

func TestRunnablePathRejectsTraversal(t *testing.T) {
	cfg := config.Config{DependenciesDir: t.TempDir()}
	for _, name := range []string{"", ".", "..", "../outside", filepath.Join("nested", "script")} {
		if _, err := runnablePath(cfg, name); err == nil {
			t.Errorf("runnablePath(%q) accepted an unsafe name", name)
		}
	}
}

func TestRunnablePathAddsShellSuffix(t *testing.T) {
	dir := t.TempDir()
	path, err := runnablePath(config.Config{DependenciesDir: dir}, "brew")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "brew.sh"); path != want {
		t.Fatalf("runnablePath() = %q, want %q", path, want)
	}
}
