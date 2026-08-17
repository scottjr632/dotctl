package cmds

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(profileCmd)
}

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage runnables",
	Long:  "Manage runnables",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
