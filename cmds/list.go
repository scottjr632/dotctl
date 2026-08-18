package cmds

import (
	"fmt"

	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all tracked files in the dotfiles repository",
	Long:    "List all tracked files in the dotfiles repository",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := inspectionConfig()
		if err != nil {
			return err
		}
		files, err := git.ListTrackedFiles(cfg).Unwrap()
		if err != nil {
			return err
		}
		if wantsJSON(cmd) {
			return writeJSON(cmd, files)
		}
		for _, file := range files {
			fmt.Fprintln(cmd.OutOrStdout(), file)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
