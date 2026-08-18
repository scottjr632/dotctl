package cmds

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/srctl/dotctl/internal/config"
	"github.com/srctl/dotctl/internal/git"
	"github.com/srctl/dotctl/internal/profile"
)

var (
	checkoutBackupExisting bool
	checkoutBackupDir      string
	checkoutSkipProfile    bool
)

type checkoutResult struct {
	BackupDir string       `json:"backup_dir,omitempty"`
	BackedUp  []string     `json:"backed_up"`
	Profile   *applyResult `json:"profile,omitempty"`
}

var checkoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "Check out tracked dotfiles into the work tree",
	Long:  "Check out tracked dotfiles into the work tree, optionally backing up paths that would block a first checkout",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkoutBackupDir != "" && !checkoutBackupExisting {
			return &cliError{code: "INVALID_ARGUMENT", err: errors.New("--backup-dir requires --backup-existing")}
		}

		cfg, err := commandConfig()
		if err != nil {
			return err
		}
		if !checkoutBackupExisting {
			if dryRun {
				actions := append([]planAction{checkoutAction(cfg)}, checkoutProfileActions(cfg)...)
				return writePlan(cmd, actions...)
			}
			if err := git.GitCmd(cfg, "checkout").ExecuteInTerminal(); err != nil {
				return fmt.Errorf("checkout failed; resolve conflicting files before retrying: %w", err)
			}
			result := checkoutResult{BackedUp: []string{}}
			result.Profile = checkoutApplyProfile(cmd, cfg)
			if jsonOutput {
				return writeJSON(cmd, result)
			}
			return nil
		}

		indexed, err := git.ListTrackedFiles(cfg).Unwrap()
		if err != nil {
			return err
		}
		if len(indexed) != 0 {
			return &cliError{
				code: "INVALID_ARGUMENT",
				err:  errors.New("--backup-existing is only available before the first checkout"),
			}
		}
		tracked, err := git.ListHeadFiles(cfg).Unwrap()
		if err != nil {
			return err
		}
		collisions, err := existingCheckoutPaths(git.WorkTree(), tracked)
		if err != nil {
			return err
		}

		if dryRun {
			actions := []planAction{}
			if len(collisions) != 0 {
				location := checkoutBackupDir
				if location == "" {
					location = filepath.Join(config.DirPath(), "backups", "<timestamp>")
				}
				actions = append(actions, action("backup_existing", fmt.Sprintf("Move %d existing path(s) to %s", len(collisions), location)))
			}
			actions = append(actions, checkoutAction(cfg))
			actions = append(actions, checkoutProfileActions(cfg)...)
			return writePlan(cmd, actions...)
		}

		result := checkoutResult{BackedUp: collisions}
		if len(collisions) != 0 {
			result.BackupDir = checkoutBackupDir
			if result.BackupDir == "" {
				result.BackupDir = filepath.Join(config.DirPath(), "backups", time.Now().UTC().Format("20060102T150405.000000000Z"))
			}
			if err := backupCheckoutPaths(git.WorkTree(), result.BackupDir, collisions); err != nil {
				return fmt.Errorf("back up existing paths: %w", err)
			}
		}

		if err := git.GitCmd(cfg, "checkout").ExecuteInTerminal(); err != nil {
			if result.BackupDir != "" {
				return fmt.Errorf("checkout failed after existing paths were backed up to %s: %w", result.BackupDir, err)
			}
			return fmt.Errorf("checkout failed: %w", err)
		}
		result.Profile = checkoutApplyProfile(cmd, cfg)
		if jsonOutput {
			return writeJSON(cmd, result)
		}
		if result.BackupDir != "" {
			color.New(color.FgGreen).Fprintf(cmd.OutOrStdout(), "Backed up %d existing path(s) to %s\n", len(result.BackedUp), result.BackupDir)
		}
		color.New(color.FgGreen).Fprintln(cmd.OutOrStdout(), "Successfully checked out dotfiles")
		return nil
	},
}

func checkoutAction(cfg config.Config) planAction {
	return action("checkout", fmt.Sprintf("Check out tracked files from %s into %s", cfg.DotfilesGitPath, git.WorkTree()))
}

// checkoutProfileActions previews the variant links that follow a checkout. The
// repository is not checked out yet during a dry run, so this reports the
// variants recorded in HEAD rather than the ones present in the work tree.
func checkoutProfileActions(cfg config.Config) []planAction {
	if checkoutSkipProfile {
		return nil
	}
	tracked, err := git.ListHeadFiles(cfg).Unwrap()
	if err != nil {
		return nil
	}
	actions := []planAction{}
	for _, target := range profile.Resolve(tracked, profile.Detect(cfg.Profile)).Targets {
		if target.Source != "" {
			actions = append(actions, action("link", fmt.Sprintf("Link %s to %s", target.Path, target.Source)))
		}
	}
	return actions
}

// checkoutApplyProfile links variants after a successful checkout. Checkout has
// already succeeded here, so an unresolved variant is reported as a warning
// instead of failing the command.
func checkoutApplyProfile(cmd *cobra.Command, cfg config.Config) *applyResult {
	if checkoutSkipProfile {
		return nil
	}
	result, err := planProfileLinks(cfg)
	if err == nil {
		err = applyProfileLinks(result)
	}
	if err != nil {
		result.Conflicts = append(result.Conflicts, linkAction{Status: linkStatusConflict, Reason: err.Error()})
	}
	if len(result.Links) == 0 && len(result.Conflicts) == 0 {
		return nil
	}
	if !jsonOutput {
		reportProfileLinks(cmd, result)
	}
	return &result
}

func existingCheckoutPaths(workTree string, tracked []string) ([]string, error) {
	tracked = append([]string(nil), tracked...)
	sort.Strings(tracked)
	selected := map[string]bool{}
	paths := []string{}

	for _, trackedPath := range tracked {
		rel := filepath.Clean(filepath.FromSlash(trackedPath))
		if rel == "." || filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("invalid tracked path %q", trackedPath)
		}

		parts := strings.Split(rel, string(filepath.Separator))
		for index := range parts {
			candidate := filepath.Join(parts[:index+1]...)
			if hasSelectedAncestor(selected, candidate) {
				break
			}
			info, err := os.Lstat(filepath.Join(workTree, candidate))
			if errors.Is(err, os.ErrNotExist) {
				break
			}
			if err != nil {
				return nil, err
			}
			isTrackedPath := index == len(parts)-1
			if !isTrackedPath && info.IsDir() {
				continue
			}
			selected[candidate] = true
			paths = append(paths, filepath.ToSlash(candidate))
			break
		}
	}
	return paths, nil
}

func hasSelectedAncestor(selected map[string]bool, path string) bool {
	for candidate := path; candidate != "."; candidate = filepath.Dir(candidate) {
		if selected[candidate] {
			return true
		}
	}
	return false
}

func backupCheckoutPaths(workTree, backupDir string, paths []string) error {
	backupDir, err := filepath.Abs(backupDir)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(backupDir); err == nil {
		return fmt.Errorf("backup directory already exists: %s", backupDir)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	for _, path := range paths {
		source := filepath.Join(workTree, filepath.FromSlash(path))
		if containsPath(source, backupDir) {
			return fmt.Errorf("backup directory %s is inside path being backed up: %s", backupDir, path)
		}
	}
	if err := os.MkdirAll(backupDir, 0o700); err != nil {
		return err
	}

	moved := []string{}
	for _, path := range paths {
		rel := filepath.FromSlash(path)
		source := filepath.Join(workTree, rel)
		destination := filepath.Join(backupDir, rel)
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return rollbackCheckoutBackup(workTree, backupDir, moved, err)
		}
		if err := os.Rename(source, destination); err != nil {
			return rollbackCheckoutBackup(workTree, backupDir, moved, err)
		}
		moved = append(moved, path)
	}
	return nil
}

func rollbackCheckoutBackup(workTree, backupDir string, moved []string, originalErr error) error {
	for index := len(moved) - 1; index >= 0; index-- {
		rel := filepath.FromSlash(moved[index])
		source := filepath.Join(backupDir, rel)
		destination := filepath.Join(workTree, rel)
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return fmt.Errorf("%w; also failed to restore %s: %v", originalErr, moved[index], err)
		}
		if err := os.Rename(source, destination); err != nil {
			return fmt.Errorf("%w; also failed to restore %s: %v", originalErr, moved[index], err)
		}
	}
	_ = os.RemoveAll(backupDir)
	return originalErr
}

func containsPath(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func init() {
	checkoutCmd.Flags().BoolVar(&checkoutBackupExisting, "backup-existing", false, "back up paths that would block the first checkout")
	checkoutCmd.Flags().StringVar(&checkoutBackupDir, "backup-dir", "", "backup destination (default: CONFIG_DIR/backups/TIMESTAMP)")
	checkoutCmd.Flags().BoolVar(&checkoutSkipProfile, "skip-profile", false, "do not link per-machine variants after checking out")
	rootCmd.AddCommand(checkoutCmd)
}
