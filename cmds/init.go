package cmds

import (
	"fmt"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/srctl/dotctl/internal/config"
	"github.com/srctl/dotctl/internal/git"
	"github.com/srctl/dotctl/internal/terminalcmd"
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
			repositoryActions := []planAction{action("initialize_repository", fmt.Sprintf("Initialize a bare Git repository at %s", path))}
			if repoURL != "" {
				repositoryActions = []planAction{
					action("clone_repository", fmt.Sprintf("Clone %s as a bare Git repository at %s", repoURL, path)),
					action("configure_repository", "Configure the cloned branch to pull from and push to origin"),
				}
			}
			return writePlan(cmd, append(repositoryActions,
				action("write_config", fmt.Sprintf("Write dotctl configuration to %s", config.FilePath())),
				action("create_directory", fmt.Sprintf("Create the runnable directory under %s", config.DirPath())),
			)...)
		}
		if repoURL != "" {
			cloneCmd := terminalcmd.New("git", "clone", "--bare", repoURL, path)
			if nonInteractive {
				cloneCmd.WithEnv("GIT_TERMINAL_PROMPT=0")
			}
			if err := cloneCmd.ExecuteInTerminal(); err != nil {
				return fmt.Errorf("clone dotfiles repository: %w", err)
			}
			if result := git.ConfigureClonedRepo(config.Config{DotfilesGitPath: path}); result.IsErr() {
				return fmt.Errorf("configure cloned dotfiles repository: %w", result.Err())
			}
		} else if result := git.InitBareRepo(git.InitRepoOptions{Path: path}); result.IsErr() {
			return result.Err()
		}

		if result := config.InitializeConfigFile(path); result.IsErr() {
			return result.Err()
		}
		if !jsonOutput {
			logPrinter.Fprintln(cmd.OutOrStdout(), "Successfully initialized dotfile config")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&dotfileConfigPath, "path", "p", "", "path to use for the bare Git repository (default: WORK_TREE/.cfg/.dotfiles)")
	initCmd.Flags().StringVarP(&repoURL, "clone", "c", "", "clone a bare repository from this URL")
}
