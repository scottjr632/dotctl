package cmds

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type contractResponse struct {
	OK    bool            `json:"ok"`
	Kind  string          `json:"kind"`
	Data  json.RawMessage `json:"data"`
	Error *errorResponse  `json:"error"`
}

func TestJSONHelpContract(t *testing.T) {
	payload := runJSONCommand(t, 0, "--json", "--help")
	if !payload.OK || payload.Kind != "result" {
		t.Fatalf("unexpected help response: %+v", payload)
	}
	var help helpData
	if err := json.Unmarshal(payload.Data, &help); err != nil {
		t.Fatal(err)
	}
	for _, flag := range help.Flags {
		if flag.Name == "json" {
			return
		}
	}
	t.Fatalf("structured help omitted global flags: %+v", help)
}

func TestJSONErrorContract(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "missing")
	payload := runJSONCommand(t, 1, "--json", "--config-dir", configDir, "status")
	if payload.OK || payload.Kind != "error" || payload.Error == nil || payload.Error.Code != "CONFIG_NOT_FOUND" {
		t.Fatalf("unexpected missing-config response: %+v", payload)
	}

	configDir = filepath.Join(root, "invalid")
	if err := os.Mkdir(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	payload = runJSONCommand(t, 1, "--json", "--config-dir", configDir, "status")
	if payload.Error == nil || payload.Error.Code != "CONFIG_INVALID" {
		t.Fatalf("unexpected invalid-config response: %+v", payload)
	}

	payload = runJSONCommand(t, 1, "--json", "completion", "bash")
	if payload.Error == nil || payload.Error.Code != "JSON_UNSUPPORTED" {
		t.Fatalf("unexpected unsupported-JSON response: %+v", payload)
	}
}

func TestJSONPlanAndMutationContract(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	workTree := filepath.Join(root, "work-tree")
	if err := os.Mkdir(workTree, 0o755); err != nil {
		t.Fatal(err)
	}

	plan := runJSONCommand(t, 0, "--json", "--dry-run", "--config-dir", configDir, "--work-tree", workTree, "init")
	if !plan.OK || plan.Kind != "plan" {
		t.Fatalf("unexpected plan response: %+v", plan)
	}
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Fatalf("dry-run created the config directory: %v", err)
	}

	result := runJSONCommand(t, 0, "--json", "--config-dir", configDir, "--work-tree", workTree, "init")
	if !result.OK || result.Kind != "result" {
		t.Fatalf("unexpected init response: %+v", result)
	}

	runJSONCommand(t, 0, "--json", "--config-dir", configDir, "--work-tree", workTree, "git", "config", "user.name", "dotctl test")
	runJSONCommand(t, 0, "--json", "--config-dir", configDir, "--work-tree", workTree, "git", "config", "user.email", "dotctl@example.com")
	runJSONCommand(t, 0, "--json", "--config-dir", configDir, "--work-tree", workTree, "git", "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(workTree, "example.conf"), []byte("example\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result = runJSONCommand(t, 0, "--json", "--config-dir", configDir, "--work-tree", workTree, "track", "example.conf", "--message", "Track example")
	var commandOutput commandResult
	if err := json.Unmarshal(result.Data, &commandOutput); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(commandOutput.Output, "Track example") {
		t.Fatalf("Git output was not captured: %+v", commandOutput)
	}

	runnable := filepath.Join(configDir, "runnables", "output.sh")
	if err := os.WriteFile(runnable, []byte("#!/bin/sh\nprintf 'script output\\n'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	result = runJSONCommand(t, 0, "--json", "--config-dir", configDir, "--work-tree", workTree, "dependencies", "run", "output")
	if err := json.Unmarshal(result.Data, &commandOutput); err != nil {
		t.Fatal(err)
	}
	if commandOutput.Output != "script output" {
		t.Fatalf("script output was not captured: %+v", commandOutput)
	}

	failingRunnable := filepath.Join(configDir, "runnables", "fail.sh")
	if err := os.WriteFile(failingRunnable, []byte("#!/bin/sh\nprintf 'failure output\\n'\nexit 7\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	failure := runJSONCommand(t, 1, "--json", "--config-dir", configDir, "--work-tree", workTree, "dependencies", "run", "fail")
	if failure.Error == nil || failure.Error.Code != "EXTERNAL_COMMAND_FAILED" {
		t.Fatalf("unexpected script failure: %+v", failure)
	}
	if err := json.Unmarshal(failure.Data, &commandOutput); err != nil {
		t.Fatal(err)
	}
	if commandOutput.Output != "failure output" {
		t.Fatalf("failed script output was not captured: %+v", commandOutput)
	}
}

func runJSONCommand(t *testing.T, expectedExit int, args ...string) contractResponse {
	t.Helper()
	commandArgs := append([]string{"-test.run=TestJSONHelperProcess", "--"}, args...)
	command := exec.Command(os.Args[0], commandArgs...)
	command.Env = append(os.Environ(), "DOTCTL_JSON_HELPER=1")
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatal(err)
		}
		exitCode = exitErr.ExitCode()
	}
	if exitCode != expectedExit {
		t.Fatalf("exit code = %d, want %d; stdout=%q stderr=%q", exitCode, expectedExit, stdout.String(), stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("JSON command wrote stderr: %q", stderr.String())
	}

	decoder := json.NewDecoder(&stdout)
	var payload contractResponse
	if err := decoder.Decode(&payload); err != nil {
		t.Fatalf("decode response: %v; stdout=%q", err, stdout.String())
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		t.Fatalf("command emitted more than one JSON document: %q", stdout.String())
	}
	return payload
}

func TestJSONHelperProcess(t *testing.T) {
	if os.Getenv("DOTCTL_JSON_HELPER") != "1" {
		return
	}
	separator := 0
	for index, arg := range os.Args {
		if arg == "--" {
			separator = index + 1
			break
		}
	}
	if err := execute(os.Args[separator:], os.Stdout, os.Stderr); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
