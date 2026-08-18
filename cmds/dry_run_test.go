package cmds

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

func TestInitDryRunDoesNotCreateFiles(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	workTree := filepath.Join(root, "work-tree")
	config.SetDir(configDir)
	git.SetWorkTree(workTree)

	previousDryRun := dryRun
	previousPath := dotfileConfigPath
	previousURL := repoURL
	dryRun = true
	dotfileConfigPath = ""
	repoURL = ""
	t.Cleanup(func() {
		dryRun = previousDryRun
		dotfileConfigPath = previousPath
		repoURL = previousURL
		config.SetDir("")
		git.SetWorkTree("")
	})

	var output bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&output)
	if err := initCmd.RunE(cmd, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Dry run; no changes made") {
		t.Fatalf("missing dry-run plan: %q", output.String())
	}
	for _, path := range []string{configDir, workTree} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("dry-run created %s: %v", path, err)
		}
	}
}
