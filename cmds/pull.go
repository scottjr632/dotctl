package cmds

import (
	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull the latest data from the repository",
	Long:  "Pull the latest data from the repository",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := commandConfig()
		if err != nil {
			return err
		}
		if dryRun {
			return writePlan(cmd, "Pull remote changes into the dotfiles repository")
		}
		if err := git.GitCmd(cfg, "pull").ExecuteInTerminal(); err != nil {
			return err
		}
		color.New(color.FgGreen).Fprintln(cmd.OutOrStdout(), "Successfully pulled the latest data")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
