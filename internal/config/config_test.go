package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndPreviewDoNotWriteMissingDefaults(t *testing.T) {
	configDir := t.TempDir()
	SetDir(configDir)
	t.Cleanup(func() { SetDir("") })
	writeTestConfig(t, Config{DotfilesGitPath: "/tmp/dotfiles"})

	cfg, err := Load().Unwrap()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DependenciesDir != "" || cfg.PreRunnableFile != "" {
		t.Fatalf("Load filled defaults: %+v", cfg)
	}
	preview, err := Preview().Unwrap()
	if err != nil {
		t.Fatal(err)
	}
	if preview.DependenciesDir != defaultRunnableDirPath() || preview.PreRunnableFile != defaultPreRunnableFilePath() {
		t.Fatalf("Preview did not resolve defaults: %+v", preview)
	}
	if _, err := os.Stat(defaultRunnableDirPath()); !os.IsNotExist(err) {
		t.Fatalf("Load or Preview created the runnable directory: %v", err)
	}
}

func TestGetFillsAndPersistsMissingDefaults(t *testing.T) {
	configDir := t.TempDir()
	SetDir(configDir)
	t.Cleanup(func() { SetDir("") })
	writeTestConfig(t, Config{DotfilesGitPath: "/tmp/dotfiles"})

	cfg, err := Get().Unwrap()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DependenciesDir != filepath.Join(configDir, "runnables") {
		t.Fatalf("unexpected dependencies directory: %q", cfg.DependenciesDir)
	}
	contents, err := os.ReadFile(cfg.PreRunnableFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "#!/bin/sh\n" {
		t.Fatalf("unexpected pre-runnable contents: %q", contents)
	}

	persisted, err := Load().Unwrap()
	if err != nil {
		t.Fatal(err)
	}
	if persisted != cfg {
		t.Fatalf("persisted config differs: got %+v, want %+v", persisted, cfg)
	}
}

func writeTestConfig(t *testing.T, cfg Config) {
	t.Helper()
	file, err := os.Create(FilePath())
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(file).Encode(cfg); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}
