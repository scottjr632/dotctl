package cmds

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/srctl/dotctl/internal/git"
)

var isTrackedCmd = &cobra.Command{
	Use: "is-tracked [file]",
	Aliases: []string{
		"istracked",
		"is",
		"tracked",
		"it",
	},
	Short: "Check if a file is tracked in the dotfiles repository",
	Long:  "Check if a file is tracked in the dotfiles repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := inspectionConfig()
		if err != nil {
			return err
		}
		tracked, err := git.IsTracked(cfg, args[0]).Unwrap()
		if err != nil {
			return err
		}
		if wantsJSON(cmd) {
			return writeJSON(cmd, struct {
				Path    string `json:"path"`
				Tracked bool   `json:"tracked"`
			}{Path: args[0], Tracked: tracked})
		}
		if tracked {
			color.New(color.FgGreen).Fprintf(cmd.OutOrStdout(), "File %q is tracked\n", args[0])
		} else {
			color.New(color.FgRed).Fprintf(cmd.OutOrStdout(), "File %q is not tracked\n", args[0])
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(isTrackedCmd)
}
