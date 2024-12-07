package config

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"

	"github.com/scottjr632/dotctl/internal/result"
)

type dirResult int

const (
	cfgFileName    = "config"
	cfgFileDirName = "dotctl"

	dirExist dirResult = iota
	dirDoesNotExist
	fileExist
)

var (
	cfgDirPath  = getConfigDirPath()
	cfgFilePath = getConfigFilePath()
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

type Config struct {
	DotfilesGitPath string `json:"git_repo_path"`
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

	return result.Ok(cfg)
}
