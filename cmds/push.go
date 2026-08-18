package cmds

import (
	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push changes to the remote repository",
	Long:  "Push changes to the remote repository",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := commandConfig()
		if err != nil {
			return err
		}
		if dryRun {
			return writePlan(cmd, action("push", "Push local dotfile commits to the configured remote"))
		}
		if result := git.Push(cfg); result.IsErr() {
			return result.Err()
		}
		if !jsonOutput {
			color.New(color.FgGreen).Fprintln(cmd.OutOrStdout(), "Successfully pushed changes")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
