package cmds

import (
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Show the status of the dotfiles repository",
	Long:    "Show the status of the dotfiles repository",
	Aliases: []string{"st"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := inspectionConfig()
		if err != nil {
			return err
		}
		if wantsJSON(cmd) {
			status, err := git.GetStatus(cfg).Unwrap()
			if err != nil {
				return err
			}
			return writeJSON(cmd, status)
		}
		return git.Status(cfg).Err()
	},
}

func init() {
	addJSONFlag(statusCmd)
	rootCmd.AddCommand(statusCmd)
}
