package cmds

import (
	"github.com/spf13/cobra"
	"github.com/srctl/dotctl/internal/git"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show Git history",
	Long:  "Show Git history",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := inspectionConfig()
		if err != nil {
			return err
		}
		return git.GitCmd(cfg, append([]string{"log", "--name-only"}, args...)...).ExecuteInTerminal()
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}
