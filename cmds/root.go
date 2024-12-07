package cmds

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "dotctl",
	Short: "dotctl is a tool for managing dotfiles",
	Long:  `dotctl is a tool for managing dotfiles`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
