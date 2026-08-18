// Package profile resolves per-machine variants of tracked dotfiles.
//
// A variant is a tracked path whose base name contains "##" followed by
// comma-separated conditions, for example ".gitconfig##hostname.workbook" or
// ".zshrc##os.darwin,arch.arm64". Every condition must hold for the variant to
// match the current machine. When several variants match one target, the most
// specific one wins.
package profile

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"
)

// Separator marks the start of the condition list in a variant's base name.
const Separator = "##"

// conditionWeights ranks condition kinds from most to least specific. A
// variant's score is the sum of its condition weights, so both a longer
// condition list and a more specific kind win.
var conditionWeights = map[string]int{
	"hostname": 8,
	"profile":  4,
	"arch":     2,
	"os":       1,
}

// Selectors describes the machine that variants are resolved against.
type Selectors struct {
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Profile  string `json:"profile,omitempty"`
}

// Detect reports the selectors for the current machine. The host name is
// shortened to the segment before the first dot so that "workbook.local" and
// "workbook" select the same variants.
func Detect(profileName string) Selectors {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = ""
	}
	return Selectors{
		Hostname: strings.ToLower(strings.SplitN(hostname, ".", 2)[0]),
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		Profile:  profileName,
	}
}

func (s Selectors) value(kind string) string {
	switch kind {
	case "hostname":
		return s.Hostname
	case "os":
		return s.OS
	case "arch":
		return s.Arch
	case "profile":
		return s.Profile
	}
	return ""
}

// Candidate is one variant competing for a target path.
type Candidate struct {
	Source     string   `json:"source"`
	Conditions []string `json:"conditions"`
	Matches    bool     `json:"matches"`
	Selected   bool     `json:"selected"`
	score      int
}

// Target is a plain path together with every variant that competes for it.
type Target struct {
	Path       string      `json:"path"`
	Source     string      `json:"source,omitempty"`
	Candidates []Candidate `json:"candidates"`
}

// InvalidVariant is a path that looks like a variant but cannot be understood.
// These are reported rather than ignored, because a silently skipped variant
// looks exactly like a variant that did not match.
type InvalidVariant struct {
	Source string `json:"source"`
	Reason string `json:"reason"`
}

// Resolution is the outcome of matching every tracked variant against one machine.
type Resolution struct {
	Selectors Selectors        `json:"selectors"`
	Targets   []Target         `json:"targets"`
	Invalid   []InvalidVariant `json:"invalid"`
}

// IsVariant reports whether a tracked path names a variant.
func IsVariant(trackedPath string) bool {
	return strings.Contains(path.Base(trackedPath), Separator)
}

// parseVariant splits a variant path into its target path and conditions.
func parseVariant(trackedPath string) (string, []string, error) {
	dir, base := path.Split(trackedPath)
	name, conditions, _ := strings.Cut(base, Separator)
	if name == "" {
		return "", nil, fmt.Errorf("variant has no target file name")
	}
	if strings.TrimSpace(conditions) == "" {
		return "", nil, fmt.Errorf("variant has no conditions after %q", Separator)
	}
	return dir + name, strings.Split(conditions, ","), nil
}

func scoreConditions(conditions []string, selectors Selectors) (matches bool, score int, err error) {
	matches = true
	for _, condition := range conditions {
		kind, value, found := strings.Cut(condition, ".")
		if !found || value == "" {
			return false, 0, fmt.Errorf("condition %q is not in kind.value form", condition)
		}
		weight, known := conditionWeights[kind]
		if !known {
			return false, 0, fmt.Errorf("unknown condition kind %q (want %s)", kind, strings.Join(knownKinds(), ", "))
		}
		score += weight
		if !strings.EqualFold(selectors.value(kind), value) {
			matches = false
		}
	}
	return matches, score, nil
}

func knownKinds() []string {
	kinds := make([]string, 0, len(conditionWeights))
	for kind := range conditionWeights {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	return kinds
}

// Resolve groups tracked variants by target and picks a winner for each. The
// returned targets and their candidates are ordered deterministically.
func Resolve(trackedPaths []string, selectors Selectors) Resolution {
	resolution := Resolution{Selectors: selectors, Targets: []Target{}, Invalid: []InvalidVariant{}}
	byTarget := map[string][]Candidate{}

	for _, trackedPath := range trackedPaths {
		if !IsVariant(trackedPath) {
			continue
		}
		targetPath, conditions, err := parseVariant(trackedPath)
		if err == nil {
			var matches bool
			var score int
			matches, score, err = scoreConditions(conditions, selectors)
			if err == nil {
				byTarget[targetPath] = append(byTarget[targetPath], Candidate{
					Source:     trackedPath,
					Conditions: conditions,
					Matches:    matches,
					score:      score,
				})
				continue
			}
		}
		resolution.Invalid = append(resolution.Invalid, InvalidVariant{Source: trackedPath, Reason: err.Error()})
	}

	for targetPath, candidates := range byTarget {
		// Highest score first, then source path, so a tie never resolves at random.
		sort.Slice(candidates, func(i, j int) bool {
			if candidates[i].score != candidates[j].score {
				return candidates[i].score > candidates[j].score
			}
			return candidates[i].Source < candidates[j].Source
		})
		target := Target{Path: targetPath, Candidates: candidates}
		for index, candidate := range candidates {
			if candidate.Matches {
				target.Candidates[index].Selected = true
				target.Source = candidate.Source
				break
			}
		}
		resolution.Targets = append(resolution.Targets, target)
	}

	sort.Slice(resolution.Targets, func(i, j int) bool {
		return resolution.Targets[i].Path < resolution.Targets[j].Path
	})
	sort.Slice(resolution.Invalid, func(i, j int) bool {
		return resolution.Invalid[i].Source < resolution.Invalid[j].Source
	})
	return resolution
}
