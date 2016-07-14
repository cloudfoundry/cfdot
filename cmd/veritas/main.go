package main

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/veritas/commands"

	"github.com/jessevdk/go-flags"
)

func main() {
	bbsParser := flags.NewParser(&commands.BBSOptions, flags.IgnoreUnknown|flags.PassDoubleDash)
	// ignoring error since we catch on the main parser below
	bbsParser.Parse()
	bbsClient := bbs.NewClient(commands.BBSOptions.BBSURL)

	parser := flags.NewParser(&commands.Veritas, flags.HelpFlag|flags.PassDoubleDash)
	logger := lager.NewLogger("veritas")
	commands.Configure(logger, os.Stdout, bbsClient)

	_, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
