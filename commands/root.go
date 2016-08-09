package commands

import (
	"code.cloudfoundry.org/lager"
	"github.com/spf13/cobra"
)

var globalLogger = lager.NewLogger("cfdot")

var RootCmd = &cobra.Command{
	Use:   "cfdot",
	Short: "Diego operator tooling",
	Long:  "A command-line tool to interact with a Cloud Foundry Diego deployment",
}

type CFDotError struct {
	message  string
	exitCode int
}

func (a CFDotError) Error() string {
	return a.message
}

func (a CFDotError) ExitCode() int {
	return a.exitCode
}
