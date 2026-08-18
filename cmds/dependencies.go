package cmds

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/srctl/dotctl/internal/config"
	"github.com/srctl/dotctl/internal/runnables"
	"github.com/srctl/dotctl/internal/utils"
)

var (
	filter    string
	withPre   bool
	newNoEdit bool
)

var dependenciesCmd = &cobra.Command{
	Use:   "dependencies",
	Short: "Create or run files that install system dependencies",
	Aliases: []string{
		"deps",
		"dep",
		"dp",
	},
	RunE: listRunnables,
}

var listAsStringsCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all runnables",
	Long:    "List all runnables",
	RunE:    listRunnables,
}

func listRunnables(cmd *cobra.Command, args []string) error {
	cfg, err := inspectionConfig()
	if err != nil {
		return err
	}
	names, err := runnables.ListAllRunnablesAsStrings(cfg).Unwrap()
	if err != nil {
		return err
	}
	if wantsJSON(cmd) {
		return writeJSON(cmd, names)
	}
	if len(names) == 0 {
		return fmt.Errorf("no runnables found")
	}
	for _, name := range names {
		fmt.Fprintln(cmd.OutOrStdout(), "* "+name)
	}
	return nil
}

var newCmd = &cobra.Command{
	Use:     "new [name]",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"n"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := commandConfig()
		if err != nil {
			return err
		}
		if dryRun {
			actions := []planAction{action("create_runnable", fmt.Sprintf("Create runnable %s in %s", args[0], cfg.DependenciesDir))}
			if !newNoEdit && !nonInteractive {
				actions = append(actions, action("open_editor", "Open the new runnable in the configured editor"))
			}
			return writePlan(cmd, actions...)
		}
		if newNoEdit || nonInteractive {
			return runnables.CreateRunnable(cfg, args[0]).Err()
		}
		return runnables.CreateNewRunnable(cfg, args[0]).Err()
	},
}

var editCmd = &cobra.Command{
	Use:   "edit [name]",
	Short: "Edit a runnable",
	Long:  "Edit a runnable",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if nonInteractive && len(args) == 0 {
			return fmt.Errorf("a runnable name is required in non-interactive mode")
		}
		cfg, err := commandConfig()
		if err != nil {
			return err
		}
		name, err := chooseRunnable(cfg, args, "Select a runnable")
		if err != nil {
			return err
		}
		if nonInteractive {
			return fmt.Errorf("edit is unavailable in non-interactive mode")
		}
		if dryRun {
			return writePlan(cmd, action("open_editor", fmt.Sprintf("Open runnable %s in the configured editor, creating it if needed", name)))
		}
		return runnables.EditRunnable(cfg, name).Err()
	},
}

var deleteCmd = &cobra.Command{
	Use:     "delete [name]",
	Short:   "Delete a runnable",
	Long:    "Delete a runnable",
	Aliases: []string{"del", "d"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if nonInteractive && len(args) == 0 {
			return fmt.Errorf("a runnable name is required in non-interactive mode")
		}
		cfg, err := commandConfig()
		if err != nil {
			return err
		}
		name, err := chooseRunnable(cfg, args, "Select a runnable")
		if err != nil {
			return err
		}
		if dryRun {
			return writePlan(cmd, action("delete_runnable", fmt.Sprintf("Delete runnable %s from %s", name, cfg.DependenciesDir)))
		}
		if nonInteractive && !assumeYes {
			return fmt.Errorf("--yes is required to delete a runnable in non-interactive mode")
		}
		return runnables.DeleteRunnable(cfg, name, assumeYes).Err()
	},
}

var runnableCmd = &cobra.Command{
	Use:   "run [name]",
	Short: "Run a runnable",
	Long:  "Run a runnable",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := commandConfig()
		if err != nil {
			return err
		}
		if nonInteractive && len(args) == 0 {
			return fmt.Errorf("a runnable name is required in non-interactive mode")
		}
		name, err := chooseRunnable(cfg, args, "Select a runnable")
		if err != nil {
			return err
		}
		if dryRun {
			actions := []planAction{}
			if withPre {
				actions = append(actions, action("execute_runnable", fmt.Sprintf("Execute pre-runnable %s", cfg.PreRunnableFile)))
			}
			actions = append(actions, action("execute_runnable", fmt.Sprintf("Execute runnable %s from %s", name, cfg.DependenciesDir)))
			return writePlan(cmd, actions...)
		}
		if withPre {
			if result := runnables.RunPreRunnable(cfg); result.IsErr() {
				return result.Err()
			}
		}
		return runnables.RunRunnable(cfg, name).Err()
	},
}

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Run all runnables",
	Long:  "Run all runnables",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := commandConfig()
		if err != nil {
			return err
		}
		names, err := runnables.ListAllRunnablesAsStrings(cfg).Unwrap()
		if err != nil {
			return err
		}
		names = utils.WithoutStrings(utils.FilterStrings(names, filter), []string{"pre", "pre.sh"})
		if !jsonOutput {
			for _, name := range names {
				fmt.Fprintln(cmd.OutOrStdout(), "* "+name)
			}
		}
		if dryRun {
			actions := make([]planAction, 0, len(names)+1)
			if withPre {
				actions = append(actions, action("execute_runnable", fmt.Sprintf("Execute pre-runnable %s", cfg.PreRunnableFile)))
			}
			for _, name := range names {
				actions = append(actions, action("execute_runnable", fmt.Sprintf("Execute runnable %s from %s", name, cfg.DependenciesDir)))
			}
			if len(actions) == 0 {
				actions = append(actions, action("noop", "No runnable scripts match the request"))
			}
			return writePlan(cmd, actions...)
		}
		if nonInteractive && !assumeYes {
			return fmt.Errorf("--yes is required to run all dependencies in non-interactive mode")
		}
		if !assumeYes {
			prompt := promptui.Prompt{Label: "Are you sure you want to run the above runnables?", IsConfirm: true}
			answer, err := prompt.Run()
			if err != nil {
				return err
			}
			if answer == "" || answer == "n" || answer == "false" {
				return fmt.Errorf("user declined to run runnables")
			}
		}
		if withPre {
			if result := runnables.RunPreRunnable(cfg); result.IsErr() {
				return result.Err()
			}
		}
		for _, name := range names {
			if result := runnables.RunRunnable(cfg, name); result.IsErr() {
				return result.Err()
			}
		}
		return nil
	},
}

func chooseRunnable(cfg config.Config, args []string, label string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}
	options, err := runnables.ListAllRunnablesAsStrings(cfg).Unwrap()
	if err != nil {
		return "", err
	}
	prompt := promptui.Select{
		Label:             label,
		Items:             options,
		StartInSearchMode: true,
		Searcher: func(input string, index int) bool {
			return strings.Contains(options[index], input)
		},
	}
	_, name, err := prompt.Run()
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", fmt.Errorf("no runnable selected")
	}
	return name, nil
}

func init() {
	newCmd.Flags().BoolVar(&newNoEdit, "no-edit", false, "create the runnable without opening an editor")
	allCmd.Flags().StringVarP(&filter, "filter", "f", "", "filter runnables")
	allCmd.Flags().BoolVarP(&withPre, "with-pre", "p", false, "include the pre runnable")
	runnableCmd.Flags().BoolVarP(&withPre, "with-pre", "p", false, "include the pre runnable")

	dependenciesCmd.AddCommand(allCmd, runnableCmd, listAsStringsCmd, deleteCmd, newCmd, editCmd)
	rootCmd.AddCommand(dependenciesCmd)
}
