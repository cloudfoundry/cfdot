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
	var bbsClient bbs.Client
	if commands.BBSOptions.BBSURL != "" {
		bbsClient = bbs.NewClient(commands.BBSOptions.BBSURL)
	}

	parser := flags.NewParser(&commands.Veritas, flags.HelpFlag|flags.PassDoubleDash)
	logger := lager.NewLogger("veritas")
	commands.Configure(logger, os.Stdout, bbsClient)

	retargs, err := parser.Parse()
	if err != nil {
		if err == commands.ErrShowHelpMessage || (len(retargs) == 1 && retargs[0] == "") {
			helpParser := flags.NewParser(&commands.Veritas, flags.IgnoreUnknown|flags.HelpFlag)
			helpParser.NamespaceDelimiter = "-"
			helpParser.ParseArgs([]string{"-h"})
			helpParser.WriteHelp(os.Stdout)
			os.Exit(0)
		} else {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
	}
}
