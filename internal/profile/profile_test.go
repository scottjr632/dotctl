package profile

import (
	"reflect"
	"testing"
)

var testSelectors = Selectors{Hostname: "workbook", OS: "darwin", Arch: "arm64", Profile: "work"}

func TestResolveSelectsMostSpecificVariant(t *testing.T) {
	tests := []struct {
		name    string
		tracked []string
		want    string
	}{
		{
			name:    "hostname outranks a longer but weaker condition list",
			tracked: []string{".gitconfig##hostname.workbook", ".gitconfig##os.darwin,arch.arm64"},
			want:    ".gitconfig##hostname.workbook",
		},
		{
			name:    "adding a condition to the strongest kind wins",
			tracked: []string{".gitconfig##hostname.workbook", ".gitconfig##hostname.workbook,os.darwin"},
			want:    ".gitconfig##hostname.workbook,os.darwin",
		},
		{
			name:    "every condition must match",
			tracked: []string{".gitconfig##hostname.workbook,os.linux", ".gitconfig##os.darwin"},
			want:    ".gitconfig##os.darwin",
		},
		{
			name:    "a stronger kind outranks two weaker ones",
			tracked: []string{".gitconfig##profile.work", ".gitconfig##arch.arm64,os.darwin"},
			want:    ".gitconfig##profile.work",
		},
		{
			name:    "equal scores resolve by source path, not map order",
			tracked: []string{".gitconfig##os.darwin,arch.arm64", ".gitconfig##arch.arm64,os.darwin"},
			want:    ".gitconfig##arch.arm64,os.darwin",
		},
		{
			name:    "conditions match case-insensitively",
			tracked: []string{".gitconfig##hostname.WorkBook"},
			want:    ".gitconfig##hostname.WorkBook",
		},
		{
			name:    "no matching variant leaves the target unselected",
			tracked: []string{".gitconfig##hostname.laptop"},
			want:    "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolution := Resolve(test.tracked, testSelectors)
			if len(resolution.Targets) != 1 {
				t.Fatalf("targets = %+v, want one target", resolution.Targets)
			}
			if got := resolution.Targets[0].Source; got != test.want {
				t.Fatalf("selected %q, want %q", got, test.want)
			}
			if got := resolution.Targets[0].Path; got != ".gitconfig" {
				t.Fatalf("target path = %q, want .gitconfig", got)
			}
		})
	}
}

func TestResolveKeepsVariantsInNestedDirectories(t *testing.T) {
	resolution := Resolve([]string{".config/nvim/init.lua##os.darwin", ".config/nvim/init.lua"}, testSelectors)
	if len(resolution.Targets) != 1 {
		t.Fatalf("targets = %+v, want one target", resolution.Targets)
	}
	if got := resolution.Targets[0].Path; got != ".config/nvim/init.lua" {
		t.Fatalf("target path = %q", got)
	}
}

// A variant that cannot be parsed must be reported, because a silently skipped
// variant is indistinguishable from one that did not match this machine.
func TestResolveReportsUnusableVariants(t *testing.T) {
	resolution := Resolve([]string{".gitconfig##osx.darwin", ".zshrc##", ".vimrc##os"}, testSelectors)
	if len(resolution.Targets) != 0 {
		t.Fatalf("targets = %+v, want none", resolution.Targets)
	}
	sources := []string{}
	for _, invalid := range resolution.Invalid {
		if invalid.Reason == "" {
			t.Fatalf("invalid variant %q has no reason", invalid.Source)
		}
		sources = append(sources, invalid.Source)
	}
	if !reflect.DeepEqual(sources, []string{".gitconfig##osx.darwin", ".vimrc##os", ".zshrc##"}) {
		t.Fatalf("invalid variants = %#v", sources)
	}
}
