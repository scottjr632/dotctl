package cmds

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/scottjr632/dotctl/internal/profile"
	"github.com/spf13/cobra"
)

var (
	profileForce       bool
	profileNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)
)

// link statuses reported by profile apply.
const (
	linkStatusCreate   = "create"
	linkStatusRelink   = "relink"
	linkStatusCurrent  = "current"
	linkStatusConflict = "conflict"
)

type linkAction struct {
	Target string `json:"target"`
	Source string `json:"source"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type applyResult struct {
	Selectors profile.Selectors        `json:"selectors"`
	Links     []linkAction             `json:"links"`
	Conflicts []linkAction             `json:"conflicts"`
	Invalid   []profile.InvalidVariant `json:"invalid"`
}

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Select per-machine variants of tracked dotfiles",
	Long: "Select per-machine variants of tracked dotfiles.\n\n" +
		"A variant is a tracked file whose name contains " + profile.Separator + " followed by " +
		"comma-separated conditions, such as .gitconfig" + profile.Separator + "hostname.workbook or " +
		".zshrc" + profile.Separator + "os.darwin,arch.arm64. Conditions match on hostname, os, arch, " +
		"and the configured profile name. Apply links each plain path to its most specific matching variant.",
}

var profileShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the selectors used to match variants on this machine",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		selectors, err := currentSelectors()
		if err != nil {
			return err
		}
		if wantsJSON(cmd) {
			return writeJSON(cmd, selectors)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "hostname: %s\nos:       %s\narch:     %s\n", selectors.Hostname, selectors.OS, selectors.Arch)
		if selectors.Profile == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "profile:  <unset>")
			return nil
		}
		fmt.Fprintf(cmd.OutOrStdout(), "profile:  %s\n", selectors.Profile)
		return nil
	},
}

var profileSetCmd = &cobra.Command{
	Use:   "set [name]",
	Short: "Set the profile name used by profile.NAME conditions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if !profileNamePattern.MatchString(name) {
			return &cliError{
				code: "INVALID_ARGUMENT",
				err:  fmt.Errorf("profile name %q must contain only letters, digits, hyphens, and underscores", name),
			}
		}
		if dryRun {
			return writePlan(cmd, action("set_profile", fmt.Sprintf("Set profile to %q in %s", name, config.FilePath())))
		}
		cfg, err := config.SetProfile(name).Unwrap()
		if err != nil {
			return err
		}
		return reportProfileName(cmd, cfg.Profile)
	},
}

var profileUnsetCmd = &cobra.Command{
	Use:   "unset",
	Short: "Clear the configured profile name",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if dryRun {
			return writePlan(cmd, action("set_profile", fmt.Sprintf("Clear the profile in %s", config.FilePath())))
		}
		if _, err := config.SetProfile("").Unwrap(); err != nil {
			return err
		}
		return reportProfileName(cmd, "")
	},
}

var profileListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List tracked variants and the one selected for each path",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := inspectionConfig()
		if err != nil {
			return err
		}
		resolution, err := resolveVariants(cfg)
		if err != nil {
			return err
		}
		if wantsJSON(cmd) {
			return writeJSON(cmd, resolution)
		}
		if len(resolution.Targets) == 0 && len(resolution.Invalid) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No tracked variants.")
			return nil
		}
		for _, target := range resolution.Targets {
			fmt.Fprintln(cmd.OutOrStdout(), target.Path)
			for _, candidate := range target.Candidates {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s %s [%s]\n",
					candidateMarker(candidate), candidate.Source, strings.Join(candidate.Conditions, ","))
			}
		}
		for _, invalid := range resolution.Invalid {
			color.New(color.FgYellow).Fprintf(cmd.OutOrStdout(), "invalid %s: %s\n", invalid.Source, invalid.Reason)
		}
		return nil
	},
}

var profileApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Link each path to its selected variant",
	Long:  "Link each plain path in the work tree to the most specific tracked variant that matches this machine",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := commandConfig()
		if err != nil {
			return err
		}
		result, err := planProfileLinks(cfg)
		if err != nil {
			return err
		}
		if dryRun {
			return writePlan(cmd, profileActions(result)...)
		}
		if err := checkProfileConflicts(result); err != nil {
			return err
		}
		if err := applyProfileLinks(result); err != nil {
			return err
		}
		if wantsJSON(cmd) {
			return writeJSON(cmd, result)
		}
		reportProfileLinks(cmd, result)
		return nil
	},
}

func currentSelectors() (profile.Selectors, error) {
	cfg, err := inspectionConfig()
	if err != nil {
		return profile.Selectors{}, err
	}
	return profile.Detect(cfg.Profile), nil
}

func resolveVariants(cfg config.Config) (profile.Resolution, error) {
	tracked, err := git.ListTrackedFiles(cfg).Unwrap()
	if err != nil {
		return profile.Resolution{}, err
	}
	return profile.Resolve(tracked, profile.Detect(cfg.Profile)), nil
}

// planProfileLinks decides what each selected variant needs without changing anything.
func planProfileLinks(cfg config.Config) (applyResult, error) {
	tracked, err := git.ListTrackedFiles(cfg).Unwrap()
	if err != nil {
		return applyResult{}, err
	}
	trackedPaths := map[string]bool{}
	for _, trackedPath := range tracked {
		trackedPaths[trackedPath] = true
	}

	resolution := profile.Resolve(tracked, profile.Detect(cfg.Profile))
	result := applyResult{
		Selectors: resolution.Selectors,
		Links:     []linkAction{},
		Conflicts: []linkAction{},
		Invalid:   resolution.Invalid,
	}
	for _, target := range resolution.Targets {
		if target.Source == "" {
			continue
		}
		link := inspectLinkTarget(target.Path, target.Source, trackedPaths[target.Path])
		if link.Status == linkStatusConflict {
			result.Conflicts = append(result.Conflicts, link)
			continue
		}
		result.Links = append(result.Links, link)
	}
	return result, nil
}

// inspectLinkTarget classifies what applying one variant would do to the work tree.
func inspectLinkTarget(targetPath, sourcePath string, targetIsTracked bool) linkAction {
	link := linkAction{Target: targetPath, Source: sourcePath, Status: linkStatusCreate}
	conflict := func(reason string) linkAction {
		link.Status = linkStatusConflict
		link.Reason = reason
		return link
	}

	// The target must not be tracked itself: replacing it with a link would
	// leave the repository permanently dirty, which --force must not do.
	if targetIsTracked {
		return conflict(fmt.Sprintf("%s is tracked; untrack it so the variant can own the path", targetPath))
	}

	absolute := filepath.Join(git.WorkTree(), filepath.FromSlash(targetPath))
	info, err := os.Lstat(absolute)
	if errors.Is(err, os.ErrNotExist) {
		return link
	}
	if err != nil {
		return conflict(err.Error())
	}
	if info.Mode()&os.ModeSymlink != 0 {
		existing, err := os.Readlink(absolute)
		if err != nil {
			return conflict(err.Error())
		}
		if existing == path.Base(sourcePath) {
			link.Status = linkStatusCurrent
			return link
		}
		if strings.HasPrefix(existing, path.Base(targetPath)+profile.Separator) {
			link.Status = linkStatusRelink
			return link
		}
		return conflict(fmt.Sprintf("%s links to %s, which dotctl does not manage", targetPath, existing))
	}
	if info.IsDir() {
		return conflict(fmt.Sprintf("%s is a directory", targetPath))
	}
	if !profileForce {
		return conflict(fmt.Sprintf("%s already exists; pass --force to replace it", targetPath))
	}
	link.Status = linkStatusRelink
	return link
}

func checkProfileConflicts(result applyResult) error {
	if len(result.Conflicts) == 0 {
		return nil
	}
	reasons := make([]string, 0, len(result.Conflicts))
	for _, conflict := range result.Conflicts {
		reasons = append(reasons, conflict.Reason)
	}
	return &cliError{
		code: "PROFILE_CONFLICT",
		data: result,
		err:  fmt.Errorf("cannot apply %d variant(s): %s", len(result.Conflicts), strings.Join(reasons, "; ")),
	}
}

// applyProfileLinks writes the links planned by planProfileLinks. The link value
// is the variant's base name so the link stays valid if the directory moves.
func applyProfileLinks(result applyResult) error {
	for _, link := range result.Links {
		if link.Status == linkStatusCurrent {
			continue
		}
		absolute := filepath.Join(git.WorkTree(), filepath.FromSlash(link.Target))
		if link.Status == linkStatusRelink {
			if err := os.Remove(absolute); err != nil {
				return fmt.Errorf("replace %s: %w", link.Target, err)
			}
		}
		if err := os.Symlink(path.Base(link.Source), absolute); err != nil {
			return fmt.Errorf("link %s: %w", link.Target, err)
		}
	}
	return nil
}

func profileActions(result applyResult) []planAction {
	actions := []planAction{}
	for _, link := range result.Links {
		if link.Status == linkStatusCurrent {
			continue
		}
		actions = append(actions, action("link", fmt.Sprintf("Link %s to %s", link.Target, link.Source)))
	}
	for _, conflict := range result.Conflicts {
		actions = append(actions, action("conflict", fmt.Sprintf("Skip %s: %s", conflict.Target, conflict.Reason)))
	}
	if len(actions) == 0 {
		actions = append(actions, action("link", "No variants need linking"))
	}
	return actions
}

func reportProfileLinks(cmd *cobra.Command, result applyResult) {
	linked := 0
	for _, link := range result.Links {
		if link.Status != linkStatusCurrent {
			linked++
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%-8s %s -> %s\n", link.Status, link.Target, link.Source)
	}
	for _, conflict := range result.Conflicts {
		color.New(color.FgYellow).Fprintf(cmd.OutOrStdout(), "skipped %s: %s\n", conflict.Target, conflict.Reason)
	}
	for _, invalid := range result.Invalid {
		color.New(color.FgYellow).Fprintf(cmd.OutOrStdout(), "invalid %s: %s\n", invalid.Source, invalid.Reason)
	}
	color.New(color.FgGreen).Fprintf(cmd.OutOrStdout(), "Applied %d link(s)\n", linked)
}

func reportProfileName(cmd *cobra.Command, name string) error {
	if wantsJSON(cmd) {
		return writeJSON(cmd, profile.Detect(name))
	}
	if name == "" {
		fmt.Fprintln(cmd.OutOrStdout(), "Cleared the profile")
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Profile set to %s\n", name)
	return nil
}

func candidateMarker(candidate profile.Candidate) string {
	switch {
	case candidate.Selected:
		return "*"
	case candidate.Matches:
		return "+"
	default:
		return "-"
	}
}

func init() {
	profileApplyCmd.Flags().BoolVar(&profileForce, "force", false, "replace existing files at variant target paths")
	profileCmd.AddCommand(profileShowCmd, profileSetCmd, profileUnsetCmd, profileListCmd, profileApplyCmd)
	rootCmd.AddCommand(profileCmd)
}
