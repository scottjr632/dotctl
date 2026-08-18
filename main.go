package main

import (
	"os"

	"github.com/srctl/dotctl/cmds"
)

func main() {
	if err := cmds.Execute(); err != nil {
		os.Exit(1)
	}
}
