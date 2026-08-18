package cmds

import (
	"fmt"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/scottjr632/dotctl/internal/terminalcmd"
	"github.com/spf13/cobra"
)

var (
	dotfileConfigPath string
	repoURL           string

	logPrinter = color.New(color.FgGreen, color.Italic)
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize dotfile config",
	Long:  "Initialize dotfile config",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		exists, err := config.DoesConfigFileExist()
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("config file already exists at %s", config.FilePath())
		}

		path := dotfileConfigPath
		if path == "" {
			path = filepath.Join(git.WorkTree(), ".cfg", ".dotfiles")
		}
		if dryRun {
			repositoryAction := fmt.Sprintf("Initialize a bare Git repository at %s", path)
			if repoURL != "" {
				repositoryAction = fmt.Sprintf("Clone %s as a bare Git repository at %s", repoURL, path)
			}
			return writePlan(cmd,
				repositoryAction,
				fmt.Sprintf("Write dotctl configuration to %s", config.FilePath()),
				fmt.Sprintf("Create the runnable directory under %s", config.DirPath()),
			)
		}
		if repoURL != "" {
			cloneCmd := terminalcmd.New("git", "clone", "--bare", repoURL, path)
			if nonInteractive {
				cloneCmd.WithEnv("GIT_TERMINAL_PROMPT=0")
			}
			if err := cloneCmd.ExecuteInTerminal(); err != nil {
				return fmt.Errorf("clone dotfiles repository: %w", err)
			}
		} else if result := git.InitBareRepo(git.InitRepoOptions{Path: path}); result.IsErr() {
			return result.Err()
		}

		if result := config.InitializeConfigFile(path); result.IsErr() {
			return result.Err()
		}
		logPrinter.Fprintln(cmd.OutOrStdout(), "Successfully initialized dotfile config")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&dotfileConfigPath, "path", "p", "", "path to use for the bare Git repository (default: WORK_TREE/.cfg/.dotfiles)")
	initCmd.Flags().StringVarP(&repoURL, "clone", "c", "", "clone a bare repository from this URL")
}
