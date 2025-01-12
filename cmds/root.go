package cmds

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/spf13/cobra"
)

var greenPrinter = color.New(color.FgGreen, color.Bold, color.Italic)

var rootCmd = &cobra.Command{
	Use:   "dotctl",
	Short: "dotctl is a tool for managing dotfiles",
	Long:  `dotctl is a tool for managing dotfiles`,
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		cfgResult := config.Get().Must()
		syncResult := git.CheckForSync(cfgResult).Must()
		if syncResult != "" {
			greenPrinter.Println(syncResult)
		}
	},
}

func Execute() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
