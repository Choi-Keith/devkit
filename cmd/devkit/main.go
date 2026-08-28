package main

import (
	"fmt"
	"os"

	"github.com/Choi-Keith/devkit/cmd/devkit/internal/cli"
	"github.com/Choi-Keith/devkit/cmd/devkit/internal/ui"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, ui.Error("%s", err))
		os.Exit(1)
	}
}
