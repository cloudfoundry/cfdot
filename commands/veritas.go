package commands

import (
	"io"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/lager"
)

type VeritasCommand struct {
	BBSURL BBSOptionsGroup `group:"BBS Options"`

	Domains DomainsCommand `command:"domains" description:"List all domains from BBS"`

	logger    lager.Logger
	output    io.Writer
	bbsClient bbs.Client
}

type BBSOptionsGroup struct {
	BBSURL string `long:"bbsURL" description:"BBS URL" env:"BBS_URL" required:"true"`
}

var Veritas VeritasCommand
var BBSOptions BBSOptionsGroup

func Configure(logger lager.Logger, output io.Writer, bbsClient bbs.Client) {
	Veritas.logger = logger
	Veritas.output = output
	Veritas.bbsClient = bbsClient
}
