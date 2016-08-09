package main

import (
	"os"

	"code.cloudfoundry.org/cfdot/commands"
)

func main() {
	if err := commands.RootCmd.Execute(); err != nil {
		if cfDotError, ok := err.(commands.CFDotError); ok {
			os.Exit(cfDotError.ExitCode())
		}

		os.Exit(-1)
	}
}
