package cmds

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/runnables"
	"github.com/scottjr632/dotctl/internal/utils"
	"github.com/spf13/cobra"
)

var filter string
var withPre bool

var dependenciesCmd = &cobra.Command{
	Use:   "dependencies",
	Short: "Create or run files that install system dependencies",
	Aliases: []string{
		"deps",
		"dep",
		"dp",
	},
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get().Must()
		res := runnables.ListAllRunnables(cfg)
		if res.IsErr() {
			errorPrinter.Println(res.Err())
		}
	},
}

var listAsStringsCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all runnables as strings",
	Long:    `List all runnables as strings`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get().Must()
		res := runnables.ListAllRunnables(cfg)
		if res.IsErr() {
			errorPrinter.Println(res.Err())
		}
	},
}

var newCmd = &cobra.Command{
	Use:  "new [name]",
	Args: cobra.ExactArgs(1),
	Aliases: []string{
		"n",
	},
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get().Must()
		res := runnables.CreateNewRunnable(cfg, args[0])
		if res.IsErr() {
			errorPrinter.Println(res.Err())
		}
	},
}

var editCmd = &cobra.Command{
	Use:   "edit [optional name]",
	Short: "Edit a runnable",
	Long:  `Edit a runnable`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get().Must()

		if len(args) == 1 {
			res := runnables.EditRunnable(cfg, args[0])
			if res.IsErr() {
				errorPrinter.Println(res.Err())
			}
			return
		}

		options := runnables.ListAllRunnablesAsStrings(cfg).Must()
		prompt := promptui.Select{
			Label:             "Select a runnable",
			Items:             options,
			StartInSearchMode: true,
			Searcher: func(input string, index int) bool {
				return strings.Contains(options[index], input)
			},
		}

		_, result, err := prompt.Run()
		if err != nil {
			errorPrinter.Println(err)
			return
		}

		if result == "" {
			errorPrinter.Println("No runnable selected")
			return
		}

		res := runnables.EditRunnable(cfg, result)
		if res.IsErr() {
			errorPrinter.Println(res.Err())
		}
	},
}

var deleteCmd = &cobra.Command{
	Use:     "delete [name]",
	Short:   "Delete a runnable",
	Long:    `Delete a runnable`,
	Aliases: []string{"del", "d"},
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get().Must()

		if len(args) >= 1 {
			res := runnables.DeleteRunnable(cfg, args[0])
			if res.IsErr() {
				errorPrinter.Println(res.Err())
			}
			return
		}

		options := runnables.ListAllRunnablesAsStrings(cfg).Must()
		prompt := promptui.Select{
			Label:             "Select a runnable",
			Items:             options,
			StartInSearchMode: true,
			Searcher: func(input string, index int) bool {
				return strings.Contains(options[index], input)
			},
		}

		_, result, err := prompt.Run()
		if err != nil {
			errorPrinter.Println(err)
			return
		}

		if result == "" {
			errorPrinter.Println("No runnable selected")
			return
		}

		res := runnables.DeleteRunnable(cfg, result)
		if res.IsErr() {
			errorPrinter.Println(res.Err())
		}
	},
}

var runnableCmd = &cobra.Command{
	Use:   "run [name]",
	Short: "Run a runnable",
	Long:  `Run a runnable`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get().Must()

		if len(args) >= 1 {
			if withPre {
				res := runnables.RunPreRunnable(cfg)
				if res.IsErr() {
					errorPrinter.Println(res.Err())
					return
				}
			}

			res := runnables.RunRunnable(cfg, args[0])
			if res.IsErr() {
				errorPrinter.Println(res.Err())
			}
			return
		}

		options := runnables.ListAllRunnablesAsStrings(cfg).Must()
		prompt := promptui.Select{
			Label:             "Select a runnable",
			Items:             options,
			StartInSearchMode: true,
			Searcher: func(input string, index int) bool {
				return strings.Contains(options[index], input)
			},
		}

		_, result, err := prompt.Run()
		if err != nil {
			errorPrinter.Println(err)
			return
		}

		if result == "" {
			errorPrinter.Println("No runnable selected")
			return
		}

		if withPre {
			res := runnables.RunPreRunnable(cfg)
			if res.IsErr() {
				errorPrinter.Println(res.Err())
				return
			}
		}

		res := runnables.RunRunnable(cfg, result)
		if res.IsErr() {
			errorPrinter.Println(res.Err())
		}
	},
}

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Run all runnables",
	Long:  `Run all runnables`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get().Must()
		allRunnables := runnables.ListAllRunnablesAsStrings(cfg).Must()
		filtered := utils.FilterStrings(allRunnables, filter)

		for _, runnable := range filtered {
			fmt.Println("* " + runnable)
		}

		prompt := promptui.Prompt{
			Label:     "Are you sure you want to run the above runnables?",
			IsConfirm: true,
		}
		res, err := prompt.Run()
		if err != nil {
			errorPrinter.Println(err)
			return
		}

		if res == "" || res == "n" || res == "false" {
			errorPrinter.Println("User declined to run runnables")
			return
		}

		if withPre {
			res := runnables.RunPreRunnable(cfg)
			if res.IsErr() {
				errorPrinter.Println(res.Err())
				return
			}
		}

		filteredWithoutPre := utils.WithoutStrings(filtered, []string{"pre", "pre.sh"})

		for _, runnable := range filteredWithoutPre {
			res := runnables.RunRunnable(cfg, runnable)
			if res.IsErr() {
				errorPrinter.Println(res.Err())
			}
		}
	},
}

func init() {
	allCmd.Flags().StringVarP(&filter, "filter", "f", "", "filter runnables")
	allCmd.Flags().BoolVarP(&withPre, "with-pre", "p", false, "include the pre runnable")

	dependenciesCmd.AddCommand(allCmd)

	runnableCmd.Flags().BoolVarP(&withPre, "with-pre", "p", false, "include the pre runnable")

	dependenciesCmd.AddCommand(runnableCmd)
	dependenciesCmd.AddCommand(listAsStringsCmd)
	dependenciesCmd.AddCommand(deleteCmd)
	dependenciesCmd.AddCommand(newCmd)
	dependenciesCmd.AddCommand(editCmd)

	rootCmd.AddCommand(dependenciesCmd)
}
