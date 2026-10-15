package main

import (
	"fmt"
	"os"

	"github.com/muhammadjihad/sow/cmd"
)

// Set by goreleaser / ldflags at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	cmd.SetVersion(version, commit, date)
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
