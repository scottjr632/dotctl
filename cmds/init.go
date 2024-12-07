package cmds

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/scottjr632/dotctl/internal/terminalcmd"
	"github.com/spf13/cobra"
)

var (
	dotfileConfigPath string
	repoUrl           string

	errorPrinter = color.New(color.FgRed, color.Bold)
	logPrinter   = color.New(color.FgGreen, color.Italic)

	defaultDotfileConfigPath = ".cfg/.dotfiles"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize dotfile config",
	Long:  `Initialize dotfile config`,
	Run: func(cmd *cobra.Command, args []string) {
		if exist, err := config.DoesConfigFileExist(); err != nil {
			errorPrinter.Println(err)
		} else if exist {
			errorPrinter.Println("config file already exists")
			return
		}

		if repoUrl != "" {
			gitCloneCmd := terminalcmd.New("git", "clone", "--bare", repoUrl, dotfileConfigPath)
			if err := gitCloneCmd.ExecuteInTerminal(); err != nil {
				errorPrinter.Println("Failed to clone dotfiles repo:", err)
				return
			}
		} else {
			if err := git.InitBareRepo(git.InitRepoOptions{Path: dotfileConfigPath}); err.IsErr() {
				errorPrinter.Println(err.Err())
				return
			}
		}

		if err := config.InitializeConfigFile(dotfileConfigPath); err.IsErr() {
			errorPrinter.Println(err.Err())
			return
		}

		logPrinter.Println("Successfully initialized dotfile config")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	homePath, err := os.UserHomeDir()
	if err != nil {
		errorPrinter.Println("Failed to get home directory:", err)
	} else {
		defaultDotfileConfigPath = fmt.Sprintf("%s/%s", homePath, defaultDotfileConfigPath)
	}

	initCmd.Flags().StringVarP(&dotfileConfigPath, "path", "p", defaultDotfileConfigPath, "config path to use for the git repo")
	initCmd.Flags().StringVarP(&repoUrl, "clone", "c", "", "clone the dotfiles repo")
}
