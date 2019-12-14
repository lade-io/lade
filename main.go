package main

import (
	"os"

	"github.com/lade-io/lade/cmd"
)

var version = "0.0.0-dev"

func init() {
	cmd.SetVersion(version)
}

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
