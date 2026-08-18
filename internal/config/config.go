package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/scottjr632/dotctl/internal/result"
)

const (
	cfgFileName            = "config"
	cfgFileDirName         = "dotctl"
	defaultRunnableDirName = "runnables"
	preRunnableFileName    = "pre.sh"
)

var (
	dirOverride      string
	ErrConfigMissing = errors.New("dotctl config not found")
	ErrConfigInvalid = errors.New("dotctl config is invalid")
)

type Config struct {
	DotfilesGitPath string `json:"git_repo_path"`
	DependenciesDir string `json:"dependencies_dir"`
	PreRunnableFile string `json:"pre_runnable_file"`
	Profile         string `json:"profile,omitempty"`
}

func SetDir(path string) {
	dirOverride = path
}

func DirPath() string {
	if dirOverride != "" {
		return dirOverride
	}
	if path := os.Getenv("DOTCTL_CONFIG_DIR"); path != "" {
		return path
	}
	if path := os.Getenv("XDG_CONFIG_HOME"); path != "" {
		return filepath.Join(path, cfgFileDirName)
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config", cfgFileDirName)
	}
	return filepath.Join(".config", cfgFileDirName)
}

func FilePath() string {
	return filepath.Join(DirPath(), cfgFileName)
}

func defaultRunnableDirPath() string {
	return filepath.Join(DirPath(), defaultRunnableDirName)
}

func defaultPreRunnableFilePath() string {
	return filepath.Join(defaultRunnableDirPath(), preRunnableFileName)
}

func allRequiredConfigsExist(cfg Config) bool {
	return cfg.DotfilesGitPath != "" && cfg.DependenciesDir != "" && cfg.PreRunnableFile != ""
}

func withDefaults(cfg Config) Config {
	if cfg.DependenciesDir == "" {
		cfg.DependenciesDir = defaultRunnableDirPath()
	}
	if cfg.PreRunnableFile == "" {
		cfg.PreRunnableFile = defaultPreRunnableFilePath()
	}
	return cfg
}

func updateMissingConfigs(cfg Config) result.Failable {
	if allRequiredConfigsExist(cfg) {
		return result.NewFailable(nil)
	}

	missingDependenciesDir := cfg.DependenciesDir == ""
	missingPreRunnable := cfg.PreRunnableFile == ""
	cfg = withDefaults(cfg)

	if missingDependenciesDir {
		if err := os.MkdirAll(cfg.DependenciesDir, 0o755); err != nil {
			return result.NewFailable(err)
		}
	}

	if missingPreRunnable {
		if err := os.MkdirAll(filepath.Dir(cfg.PreRunnableFile), 0o755); err != nil {
			return result.NewFailable(err)
		}
		file, err := os.OpenFile(cfg.PreRunnableFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o755)
		if err != nil && !errors.Is(err, os.ErrExist) {
			return result.NewFailable(err)
		}
		if err == nil {
			if _, writeErr := file.WriteString("#!/bin/sh\n"); writeErr != nil {
				file.Close()
				return result.NewFailable(writeErr)
			}
			if closeErr := file.Close(); closeErr != nil {
				return result.NewFailable(closeErr)
			}
		}
	}

	return result.NewFailable(write(cfg))
}

func write(cfg Config) error {
	if err := os.MkdirAll(DirPath(), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(FilePath(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}

func InitializeConfigFile(path string) result.Failable {
	if _, err := os.Stat(FilePath()); err == nil {
		return result.NewFailable(errors.New("config file already exists"))
	} else if !errors.Is(err, os.ErrNotExist) {
		return result.NewFailable(err)
	}

	return updateMissingConfigs(Config{DotfilesGitPath: path})
}

func DoesConfigFileExist() (bool, error) {
	info, err := os.Stat(FilePath())
	if err == nil {
		return !info.IsDir(), nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// Load reads the config without creating files or filling missing values.
func Load() result.Result[Config] {
	file, err := os.Open(FilePath())
	if errors.Is(err, os.ErrNotExist) {
		return result.Err[Config](fmt.Errorf("%w: %s", ErrConfigMissing, FilePath()))
	}
	if err != nil {
		return result.Err[Config](err)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return result.Err[Config](fmt.Errorf("%w: %v", ErrConfigInvalid, err))
	}
	return result.Ok(cfg)
}

// Preview reads the effective config without creating or changing files.
func Preview() result.Result[Config] {
	cfg, err := Load().Unwrap()
	if err != nil {
		return result.Err[Config](err)
	}
	return result.Ok(withDefaults(cfg))
}

// SetProfile stores the profile name used to select per-machine dotfile variants.
// An empty name clears it.
func SetProfile(name string) result.Result[Config] {
	cfg, err := Get().Unwrap()
	if err != nil {
		return result.Err[Config](err)
	}
	cfg.Profile = name
	if err := write(cfg); err != nil {
		return result.Err[Config](err)
	}
	return result.Ok(cfg)
}

// Get reads the config and fills defaults used by older config files.
func Get() result.Result[Config] {
	cfg, err := Load().Unwrap()
	if err != nil {
		return result.Err[Config](err)
	}
	if updateResult := updateMissingConfigs(cfg); updateResult.IsErr() {
		return result.Err[Config](updateResult.Err())
	}
	return Load()
}
