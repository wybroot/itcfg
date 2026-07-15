package main

import (
	"fmt"
	"os"

	"itcfg/agent/internal/command"
)

var (
	Version   = "0.1.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	rootCmd := command.NewRootCommand(Version, BuildTime, GitCommit)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}