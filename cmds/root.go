package cmds

import (
	"fmt"
	"os"

	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var (
	configDir      string
	workTree       string
	nonInteractive bool
	assumeYes      bool
	dryRun         bool
)

var rootCmd = &cobra.Command{
	Use:           "dotctl",
	Short:         "Manage dotfiles stored in a bare Git repository",
	Long:          "Manage dotfiles stored in a bare Git repository",
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if configDir != "" {
			config.SetDir(configDir)
		}
		if workTree != "" {
			git.SetWorkTree(workTree)
		}
		git.SetNonInteractive(nonInteractive)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", "", "directory containing the dotctl config (or DOTCTL_CONFIG_DIR)")
	rootCmd.PersistentFlags().StringVar(&workTree, "work-tree", "", "Git work tree to manage (or DOTCTL_WORK_TREE)")
	rootCmd.PersistentFlags().BoolVar(&nonInteractive, "non-interactive", false, "fail instead of prompting or opening an editor")
	rootCmd.PersistentFlags().BoolVarP(&assumeYes, "yes", "y", false, "confirm requested operations")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "describe changes without applying them")
}

func inspectionConfig() (config.Config, error) {
	return config.Preview().Unwrap()
}

func commandConfig() (config.Config, error) {
	if dryRun {
		return inspectionConfig()
	}
	return config.Get().Unwrap()
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "dotctl:", err)
		return err
	}
	return nil
}
