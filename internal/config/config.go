package config

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"

	"github.com/fatih/color"
	"github.com/scottjr632/dotctl/internal/result"
	"github.com/scottjr632/dotctl/internal/terminalcmd"
)

type dirResult int

const (
	cfgFileName    = "config"
	cfgFileDirName = "dotctl"

	defaultRunnableDirName = "runnables"

	preRunnableFileName = "pre.sh"

	dirExist dirResult = iota
	dirDoesNotExist
	fileExist
)

var (
	cfgDirPath                 = getConfigDirPath()
	cfgFilePath                = getConfigFilePath()
	defaultCfgRunnableFilePath = getDefaultRunnableDirPath()
)

func getConfigDirPath() string {
	homePath, err := os.UserHomeDir()
	if err != nil {
		slog.Error("Failed to get home directory", "error", err)
		return ".config/" + cfgFileDirName
	}
	return homePath + "/.config/" + cfgFileDirName
}

func getConfigFilePath() string {
	return getConfigDirPath() + "/" + cfgFileName
}

func getDefaultRunnableDirPath() string {
	return getConfigDirPath() + "/" + defaultRunnableDirName
}

func getDefaultPreRunnableFilePath() string {
	return getConfigDirPath() + "/" + defaultRunnableDirName + "/" + preRunnableFileName
}

type Config struct {
	DotfilesGitPath string `json:"git_repo_path"`
	DependenciesDir string `json:"dependencies_dir"`
	PreRunnableFile string `json:"pre_runnable_file"`
}

func allRequiredConfigsExist(cfg Config) bool {
	return cfg.DotfilesGitPath != "" && cfg.DependenciesDir != "" && cfg.PreRunnableFile != ""
}

func updateMissingConfigs(currentCfg Config) result.Failable {
	if allRequiredConfigsExist(currentCfg) {
		return result.NewFailable(nil)
	}

	if currentCfg.DependenciesDir == "" {
		currentCfg.DependenciesDir = getDefaultRunnableDirPath()
		err := os.MkdirAll(currentCfg.DependenciesDir, 0755)
		if err != nil {
			return result.NewFailable(err)
		}
	}

	if currentCfg.PreRunnableFile == "" {
		currentCfg.PreRunnableFile = getDefaultPreRunnableFilePath()
		file, err := os.Create(currentCfg.PreRunnableFile)
		defer file.Close()
		file.WriteString("#!/bin/sh\n")
		if err != nil {
			return result.NewFailable(err)
		}

		if err = os.Chmod(currentCfg.PreRunnableFile, 0755); err != nil {
			return result.NewFailable(err)
		}
	}

	configFile, err := os.OpenFile(cfgFilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return result.NewFailable(err)
	}
	defer configFile.Close()

	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(currentCfg); err != nil {
		return result.NewFailable(err)
	}

	return result.NewFailable(nil)
}

func doesPathExist(path string) (dirResult, error) {
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			return dirExist, nil
		} else {
			return fileExist, nil
		}
	} else if os.IsNotExist(err) {
		return dirDoesNotExist, nil
	} else {
		return -1, err
	}
}

func InitializeConfigFile(path string) result.Failable {
	slog.Info("Initializing dotfile config", "path", cfgFilePath)
	if dir, err := doesPathExist(cfgFilePath); err != nil {
		return result.NewFailable(err)
	} else if dir == dirDoesNotExist {
		if err := os.MkdirAll(cfgDirPath, 0755); err != nil {
			return result.NewFailable(err)
		}
	}

	if dir, err := doesPathExist(cfgFilePath); err != nil {
		return result.NewFailable(err)
	} else if dir == fileExist {
		return result.NewFailable(errors.New("file already exists"))
	}

	configFile, err := os.Create(cfgFilePath)
	if err != nil {
		return result.NewFailable(err)
	}
	defer configFile.Close()

	newConfig := Config{DotfilesGitPath: path}

	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(newConfig); err != nil {
		return result.NewFailable(err)
	}

	return result.NewFailable(nil)
}

func DoesConfigFileExist() (bool, error) {
	dirResult, err := doesPathExist(cfgFilePath)
	return dirResult == fileExist, err
}

func PrintConfigFile() result.Failable {
	color.Green("Printing config file: %s", cfgFilePath)
	cat := terminalcmd.New("cat", cfgFilePath)
	return result.NewFailable(cat.ExecuteInTerminal())
}

func Get() result.Result[Config] {
	var cfg Config

	file, err := os.Open(cfgFilePath)
	if err != nil {
		return result.Err[Config](err)
	}
	defer file.Close() // Ensure the file is closed after we're done

	// Decode the JSON data into the struct
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return result.Err[Config](err)
	}

	if updateMissingConfigs(cfg).IsErr() {
		return result.Err[Config](updateMissingConfigs(cfg).Err())
	}

	return result.Ok(cfg)
}
