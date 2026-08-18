package cmds

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/scottjr632/dotctl/internal/terminalcmd"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	configDir       string
	workTree        string
	nonInteractive  bool
	assumeYes       bool
	dryRun          bool
	jsonOutput      bool
	executedCommand = "dotctl"
)

var rootCmd = &cobra.Command{
	Use:           "dotctl",
	Short:         "Manage dotfiles stored in a bare Git repository",
	Long:          "Manage dotfiles stored in a bare Git repository",
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput && (cmd.Name() == "completion" || cmd.Parent() != nil && cmd.Parent().Name() == "completion") {
			return &cliError{code: "JSON_UNSUPPORTED", err: fmt.Errorf("JSON output is not supported for shell completion generation")}
		}
		if configDir != "" {
			config.SetDir(configDir)
		}
		if workTree != "" {
			git.SetWorkTree(workTree)
		}
		if jsonOutput {
			nonInteractive = true
		}
		git.SetNonInteractive(nonInteractive)
		terminalcmd.SetCaptureOutput(jsonOutput)
		executedCommand = cmd.CommandPath()
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", "", "directory containing the dotctl config (or DOTCTL_CONFIG_DIR)")
	rootCmd.PersistentFlags().StringVar(&workTree, "work-tree", "", "Git work tree to manage (or DOTCTL_WORK_TREE)")
	rootCmd.PersistentFlags().BoolVar(&nonInteractive, "non-interactive", false, "fail instead of prompting or opening an editor")
	rootCmd.PersistentFlags().BoolVarP(&assumeYes, "yes", "y", false, "confirm requested operations")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "describe changes without applying them")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output one machine-readable JSON document")

	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if !jsonOutput {
			defaultHelp(cmd, args)
			return
		}
		_ = writeJSON(cmd, commandHelp(cmd))
	})
}

type helpCommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type helpFlag struct {
	Name        string `json:"name"`
	Shorthand   string `json:"shorthand,omitempty"`
	Description string `json:"description"`
}

type helpData struct {
	Command     string        `json:"command"`
	Usage       string        `json:"usage"`
	Description string        `json:"description"`
	Commands    []helpCommand `json:"commands"`
	Flags       []helpFlag    `json:"flags"`
}

func commandHelp(cmd *cobra.Command) helpData {
	commands := make([]helpCommand, 0, len(cmd.Commands()))
	for _, child := range cmd.Commands() {
		if child.IsAvailableCommand() {
			commands = append(commands, helpCommand{Name: child.Name(), Description: child.Short})
		}
	}
	flags := []helpFlag{}
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		flags = append(flags, helpFlag{Name: flag.Name, Shorthand: flag.Shorthand, Description: flag.Usage})
	})
	description := cmd.Long
	if description == "" {
		description = cmd.Short
	}
	return helpData{
		Command:     cmd.CommandPath(),
		Usage:       cmd.UseLine(),
		Description: description,
		Commands:    commands,
		Flags:       flags,
	}
}

func inspectionConfig() (config.Config, error) {
	return config.Preview().Unwrap()
}

func commandConfig() (config.Config, error) {
	if dryRun {
		return inspectionConfig()
	}
	return config.Get().Unwrap()
}

func Execute() error {
	return execute(os.Args[1:], os.Stdout, os.Stderr)
}

func execute(args []string, stdout, stderr io.Writer) error {
	responseWritten = false
	executedCommand = "dotctl"
	terminalcmd.ResetCapturedOutput()
	terminalcmd.SetCaptureOutput(false)

	preparedArgs, err := prepareGitArgs(args)
	if err != nil {
		if jsonOutput {
			_ = writeErrorResponse(stdout, err)
		} else {
			fmt.Fprintln(stderr, "dotctl:", err)
		}
		return err
	}
	rootCmd.SetArgs(preparedArgs)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)

	err = rootCmd.Execute()
	if jsonOutput {
		if err != nil && !responseWritten {
			_ = writeErrorResponse(stdout, err)
		}
		if err == nil && !responseWritten {
			return writeCommandResult(stdout, executedCommand, terminalcmd.CapturedOutput())
		}
		return err
	}
	if err != nil {
		fmt.Fprintln(stderr, "dotctl:", err)
	}
	return err
}

// Cobra's DisableFlagParsing is required to pass arbitrary flags to Git, but it
// also skips parent flag parsing. Parse only the global flags before "git" and
// leave everything after it untouched for Git.
func prepareGitArgs(args []string) ([]string, error) {
	gitIndex := gitCommandIndex(args)
	if gitIndex <= 0 {
		return args, nil
	}
	if err := rootCmd.PersistentFlags().Parse(args[:gitIndex]); err != nil {
		return nil, err
	}
	return append([]string{"git"}, args[gitIndex+1:]...), nil
}

func gitCommandIndex(args []string) int {
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "git":
			return index
		case "--config-dir", "--work-tree":
			index++
		case "--json", "--dry-run", "--non-interactive", "--yes", "-y":
			continue
		default:
			if strings.HasPrefix(args[index], "--config-dir=") ||
				strings.HasPrefix(args[index], "--work-tree=") ||
				strings.HasPrefix(args[index], "--json=") ||
				strings.HasPrefix(args[index], "--dry-run=") ||
				strings.HasPrefix(args[index], "--non-interactive=") ||
				strings.HasPrefix(args[index], "--yes=") {
				continue
			}
			return -1
		}
	}
	return -1
}
