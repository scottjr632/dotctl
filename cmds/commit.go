package cmds

import (
	"fmt"
	"os"

	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit changes",
	Long:  `Commit changes`,
	Run: func(cmd *cobra.Command, args []string) {
		err := git.GitCmd(config.Get().Must(), "commit").ExecuteInTerminal()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
}
