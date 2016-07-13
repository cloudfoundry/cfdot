package main

import (
	"fmt"
	"os"
)

type DomainsCommand struct {
	BbsURL string `long:"bbsURL" description:"URL to communicate to BBS"`
}

var domainsCommand DomainsCommand

func (x *DomainsCommand) Execute(args []string) error {
	var bbsURL string

	if x.BbsURL != "" {
		// bbsURL from flags
		bbsURL = x.BbsURL
	} else {
		// bbsURL from env
		bbsURL = os.Getenv("BBS_URL")
	}

	if bbsURL == "" {
		fmt.Print("bbsURL not specified")
		os.Exit(1)
	}

	fmt.Printf("bbsURL: %s", bbsURL)

	return nil
}

func init() {
	parser.AddCommand("domains",
		"Checks BBS fresh domains",
		"The domains command checks bbs domains. Use -bbsURL specify the URL.",
		&domainsCommand)
}
