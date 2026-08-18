package cmds

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/scottjr632/dotctl/internal/agentskill"
	"github.com/spf13/cobra"
)

var (
	skillInstallPath string
	forceSkill       bool
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Install or print instructions for agents using dotctl",
}

var installSkillCmd = &cobra.Command{
	Use:   "install-skill",
	Short: "Install the dotctl skill for compatible agents",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := resolvedSkillInstallPath()
		if err != nil {
			return err
		}
		if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to replace symbolic link %s", path)
		} else if err != nil && !os.IsNotExist(err) {
			return err
		}
		existing, err := os.ReadFile(path)
		if err == nil && string(existing) == agentskill.Content {
			fmt.Fprintf(cmd.OutOrStdout(), "dotctl agent skill is already installed at %s\n", path)
			return nil
		}
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err == nil && !forceSkill {
			return fmt.Errorf("%s already exists; use --force to replace it", path)
		}
		if dryRun {
			return writePlan(cmd, fmt.Sprintf("Write the dotctl agent skill to %s", path))
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(agentskill.Content), 0o644); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Installed dotctl agent skill at %s\n", path)
		return nil
	},
}

var printSkillCmd = &cobra.Command{
	Use:   "print-skill",
	Short: "Print the embedded dotctl agent skill",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := fmt.Fprint(cmd.OutOrStdout(), agentskill.Content)
		return err
	},
}

func resolvedSkillInstallPath() (string, error) {
	if skillInstallPath != "" {
		return filepath.Join(skillInstallPath, "SKILL.md"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".agents", "skills", "dotctl", "SKILL.md"), nil
}

func init() {
	installSkillCmd.Flags().StringVar(&skillInstallPath, "path", "", "skill directory (default: ~/.agents/skills/dotctl)")
	installSkillCmd.Flags().BoolVar(&forceSkill, "force", false, "replace an existing skill")
	agentCmd.AddCommand(installSkillCmd, printSkillCmd)
	rootCmd.AddCommand(agentCmd)
}
