package runnables

import (
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/scottjr632/dotctl/internal/config"
	"github.com/scottjr632/dotctl/internal/git"
	"github.com/scottjr632/dotctl/internal/result"
	"github.com/scottjr632/dotctl/internal/terminalcmd"
)

func CreateNewRunnable(cfg config.Config, name string) result.Failable {
	filename := fmt.Sprintf("%s/%s.sh", cfg.DependenciesDir, name)

	if _, err := os.Stat(filename); err == nil {
		return EditRunnable(cfg, name)
	} else if os.IsNotExist(err) {
		file, err := os.Create(filename)
		defer file.Close()

		if err != nil {
			return result.NewFailable(err)
		}

		if err = os.Chmod(filename, 0755); err != nil {
			return result.NewFailable(err)
		}

		_, err = file.WriteString("#!/bin/sh\n")
		if err != nil {
			return result.NewFailable(err)
		}
	}

	if addRes := git.AddFile(cfg, filename); addRes.IsErr() {
		return result.NewFailable(addRes.Err())
	}

	return EditRunnable(cfg, name)

}

func ListAllRunnables(cfg config.Config) result.Failable {
	files, err := os.ReadDir(cfg.DependenciesDir)
	if err != nil {
		return result.NewFailable(err)
	}

	if len(files) == 0 {
		return result.NewFailable(fmt.Errorf("no runnables found"))
	}

	for _, file := range files {
		fmt.Println("* " + file.Name())
	}
	return result.NewFailable(nil)
}

func ListAllRunnablesAsStrings(cfg config.Config) result.Result[[]string] {
	files, err := os.ReadDir(cfg.DependenciesDir)
	if err != nil {
		return result.Err[[]string](err)
	}

	var runnables []string
	for _, file := range files {
		runnables = append(runnables, file.Name())
	}

	return result.Ok(runnables)
}

func EditRunnable(cfg config.Config, name string) result.Failable {
	filename := fmt.Sprintf("%s/%s", cfg.DependenciesDir, name)
	if !strings.HasSuffix(name, ".sh") {
		filename = fmt.Sprintf("%s.sh", filename)
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if newRunnableErr := CreateNewRunnable(cfg, name); newRunnableErr.IsErr() {
			return result.NewFailable(newRunnableErr.Err())
		}
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nvim"
	}

	err := terminalcmd.New(editor, filename).ExecuteInTerminal()
	if err != nil {
		return result.NewFailable(err)
	}

	return result.NewFailable(nil)
}

func DeleteRunnable(cfg config.Config, name string) result.Failable {
	filename := fmt.Sprintf("%s/%s", cfg.DependenciesDir, name)

	prompt := promptui.Prompt{
		Label:     "Are you sure you want to delete this runnable " + name,
		IsConfirm: true,
	}

	res, err := prompt.Run()
	if err != nil {
		return result.NewFailable(err)
	}

	if res == "" || res == "n" || res == "false" {
		return result.NewFailable(fmt.Errorf("user declined to delete runnable %s", name))
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return result.NewFailable(fmt.Errorf("file %s does not exist", filename))
	}

	err = os.Remove(filename)
	if err != nil {
		return result.NewFailable(err)
	}

	return result.NewFailable(nil)
}

func RunPreRunnable(cfg config.Config) result.Failable {
	if _, err := os.Stat(cfg.PreRunnableFile); os.IsNotExist(err) {
		return result.NewFailable(fmt.Errorf("file %s does not exist", cfg.PreRunnableFile))
	}

	err := terminalcmd.New(cfg.PreRunnableFile).ExecuteInTerminal()
	if err != nil {
		return result.NewFailable(err)
	}

	return result.NewFailable(nil)
}

func getNameWithSh(name string) string {
	if strings.HasSuffix(name, ".sh") {
		return name
	}
	return name + ".sh"
}

func RunRunnable(cfg config.Config, name string) result.Failable {
	nameWithSh := getNameWithSh(name)
	filename := fmt.Sprintf("%s/%s", cfg.DependenciesDir, nameWithSh)

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return result.NewFailable(fmt.Errorf("file %s does not exist", filename))
	}

	err := terminalcmd.New(filename).ExecuteInTerminal()
	if err != nil {
		return result.NewFailable(err)
	}

	return result.NewFailable(nil)
}
