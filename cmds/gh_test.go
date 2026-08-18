package cmds

import (
	"reflect"
	"testing"
)

func TestGitHubWebURL(t *testing.T) {
	tests := map[string]string{
		"git@github.com:example/repo.git":       "https://github.com/example/repo",
		"ssh://git@github.com/example/repo.git": "https://github.com/example/repo",
		"https://github.com/example/repo.git":   "https://github.com/example/repo",
		"https://github.com/example/repo":       "https://github.com/example/repo",
	}
	for remote, expected := range tests {
		actual, err := githubWebURL(remote)
		if err != nil {
			t.Fatalf("githubWebURL(%q): %v", remote, err)
		}
		if actual != expected {
			t.Fatalf("githubWebURL(%q) = %q, want %q", remote, actual, expected)
		}
	}
	for _, remote := range []string{"", "git@github.com:example", "https://example.com/example/repo", "https://evilgithub.com/example/repo"} {
		if _, err := githubWebURL(remote); err == nil {
			t.Errorf("githubWebURL(%q) accepted an invalid remote", remote)
		}
	}
}

func TestBrowserCommand(t *testing.T) {
	tests := []struct {
		goos       string
		executable string
		arguments  []string
	}{
		{goos: "darwin", executable: "open", arguments: []string{"https://github.com/example/repo"}},
		{goos: "linux", executable: "xdg-open", arguments: []string{"https://github.com/example/repo"}},
		{goos: "windows", executable: "rundll32", arguments: []string{"url.dll,FileProtocolHandler", "https://github.com/example/repo"}},
	}
	for _, test := range tests {
		t.Run(test.goos, func(t *testing.T) {
			executable, arguments, err := browserCommand(test.goos, "https://github.com/example/repo")
			if err != nil {
				t.Fatal(err)
			}
			if executable != test.executable || !reflect.DeepEqual(arguments, test.arguments) {
				t.Fatalf("browserCommand() = %q, %#v; want %q, %#v", executable, arguments, test.executable, test.arguments)
			}
		})
	}

	if _, _, err := browserCommand("plan9", "https://github.com/example/repo"); err == nil {
		t.Fatal("browserCommand() accepted an unsupported platform")
	}
}
