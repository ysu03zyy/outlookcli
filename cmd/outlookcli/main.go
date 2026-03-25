package main

import (
	"os"

	"github.com/ysu03zyy/outlookcli/internal/cmd"
)

// version is set by -ldflags at link time (e.g. -X main.version=1.0.0).
var version = "dev"

func init() {
	cmd.Version = version
}

func main() {
	if err := cmd.Execute(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
