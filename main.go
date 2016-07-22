package main

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cfdot/commands"
)

func main() {
	if err := commands.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
