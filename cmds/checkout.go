package cmds

import (
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var checkoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "Checkout a branch in the dotfiles repository",
	Long:  `Checkout a branch in the dotfiles repository`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgResult := config.Get()
		if cfgResult.IsErr() {
			color.Red("Failed to get config: %v", cfgResult.UnwrapErr())
			return
		}

		cfg := cfgResult.Must()
		checkoutCmd := git.GitCmd(cfg, "checkout")
		if err := checkoutCmd.ExecuteInTerminal(); err != nil {
			color.Red("Failed to checkout branch trying to create backup")
			createBackup(cfg)
			return
		}
	},
}

func createBackup(cfg config.Config) error {
	backupDir, err := createBackupDir()
	if err != nil {
		return err
	}

	checkoutCmd := git.GitCmd(cfg, "checkout", "2>&1 | egrep \\\"s+\\.\" | awk {'print $1'} | xargs -I{} mv {} "+backupDir+"/{}")
	checkoutCmd.ExecuteInTerminal()

	return git.GitCmd(cfg, "checkout").ExecuteInTerminal()
}

func createBackupDir() (string, error) {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	backupDir := filepath.Join(homePath, ".config-backup")
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		if err := os.MkdirAll(backupDir, 0o755); err != nil {
			return "", err
		}
	}

	return backupDir, nil
}

func init() {
	rootCmd.AddCommand(checkoutCmd)
}
