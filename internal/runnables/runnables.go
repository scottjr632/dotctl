package runnables

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/scottjr632/dotctl/internal/result"
	"github.com/scottjr632/dotctl/internal/terminalcmd"
)

func CreateRunnable(cfg config.Config, name string) result.Failable {
	filename, err := runnablePath(cfg, name)
	if err != nil {
		return result.NewFailable(err)
	}
	if _, err := os.Stat(filename); err == nil {
		return result.NewFailable(fmt.Errorf("runnable %q already exists", name))
	} else if !os.IsNotExist(err) {
		return result.NewFailable(err)
	}
	if err := os.MkdirAll(cfg.DependenciesDir, 0o755); err != nil {
		return result.NewFailable(err)
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o755)
	if err != nil {
		return result.NewFailable(err)
	}
	if _, err := file.WriteString("#!/bin/sh\n"); err != nil {
		file.Close()
		return result.NewFailable(err)
	}
	if err := file.Close(); err != nil {
		return result.NewFailable(err)
	}
	return git.AddFile(cfg, filename)
}

func CreateNewRunnable(cfg config.Config, name string) result.Failable {
	filename, err := runnablePath(cfg, name)
	if err != nil {
		return result.NewFailable(err)
	}
	if _, err := os.Stat(filename); err == nil {
		return EditRunnable(cfg, name)
	} else if !os.IsNotExist(err) {
		return result.NewFailable(err)
	}
	if result := CreateRunnable(cfg, name); result.IsErr() {
		return result
	}
	return EditRunnable(cfg, name)
}

func ListAllRunnablesAsStrings(cfg config.Config) result.Result[[]string] {
	files, err := os.ReadDir(cfg.DependenciesDir)
	if err != nil {
		return result.Err[[]string](err)
	}

	names := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sh") {
			names = append(names, file.Name())
		}
	}
	return result.Ok(names)
}

func EditRunnable(cfg config.Config, name string) result.Failable {
	filename, err := runnablePath(cfg, name)
	if err != nil {
		return result.NewFailable(err)
	}
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if result := CreateRunnable(cfg, name); result.IsErr() {
			return result
		}
	} else if err != nil {
		return result.NewFailable(err)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nvim"
	}
	return result.NewFailable(terminalcmd.New(editor, filename).ExecuteInTerminal())
}

func DeleteRunnable(cfg config.Config, name string, skipConfirmation bool) result.Failable {
	filename, err := runnablePath(cfg, name)
	if err != nil {
		return result.NewFailable(err)
	}
	if !skipConfirmation {
		prompt := promptui.Prompt{
			Label:     "Are you sure you want to delete this runnable " + name,
			IsConfirm: true,
		}
		answer, err := prompt.Run()
		if err != nil {
			return result.NewFailable(err)
		}
		if answer == "" || answer == "n" || answer == "false" {
			return result.NewFailable(fmt.Errorf("user declined to delete runnable %s", name))
		}
	}
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return result.NewFailable(fmt.Errorf("file %s does not exist", filename))
		}
		return result.NewFailable(err)
	}
	return result.NewFailable(os.Remove(filename))
}

func RunPreRunnable(cfg config.Config) result.Failable {
	if _, err := os.Stat(cfg.PreRunnableFile); err != nil {
		if os.IsNotExist(err) {
			return result.NewFailable(fmt.Errorf("file %s does not exist", cfg.PreRunnableFile))
		}
		return result.NewFailable(err)
	}
	return result.NewFailable(terminalcmd.New(cfg.PreRunnableFile).ExecuteInTerminal())
}

func RunRunnable(cfg config.Config, name string) result.Failable {
	filename, err := runnablePath(cfg, name)
	if err != nil {
		return result.NewFailable(err)
	}
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return result.NewFailable(fmt.Errorf("file %s does not exist", filename))
		}
		return result.NewFailable(err)
	}
	return result.NewFailable(terminalcmd.New(filename).ExecuteInTerminal())
}

func runnablePath(cfg config.Config, name string) (string, error) {
	if name == "" || name == "." || name == ".." || filepath.Base(name) != name {
		return "", fmt.Errorf("invalid runnable name %q", name)
	}
	if !strings.HasSuffix(name, ".sh") {
		name += ".sh"
	}
	return filepath.Join(cfg.DependenciesDir, name), nil
}
