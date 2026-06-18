package main

import (
	"os"

	"github.com/ajmandourah/bazarr-sync/internal/cli"
	"github.com/ajmandourah/bazarr-sync/internal/tui"
)

func main() {
	// TUI: launch when no arguments provided
	if len(os.Args) == 1 {
		tui.Run()
		return
	}

	// CLI: legacy behavior with arguments
	cli.Execute()
}
