package main

import (
	"os"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	// add veritas level flags here
}

var Opts Options
var parser = flags.NewParser(&Opts, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	} else {
	}
}
