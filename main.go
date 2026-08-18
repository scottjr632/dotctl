package main

import (
	"os"

	"github.com/scottjr632/dotctl/cmds"
)

func main() {
	if err := cmds.Execute(); err != nil {
		os.Exit(1)
	}
}
