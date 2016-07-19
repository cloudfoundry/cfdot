package commands

import (
	"io"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/lager"
)

type CFdotCommand struct {
	BBSOptions BBSOptionsGroup `group:"BBS Options"`
	Domains    DomainsCommand  `command:"domains" description:"List all fresh domains."`
	Help       HelpCommand     `command:"help" description:"Show usage information for commands."`

	logger    lager.Logger
	output    io.Writer
	bbsClient bbs.Client
}

type BBSOptionsGroup struct {
	BBSURL string `long:"bbsURL" description:"BBS URL" env:"BBS_URL"`
}

var CFdot CFdotCommand
var BBSOptions BBSOptionsGroup

func Configure(logger lager.Logger, output io.Writer, bbsClient bbs.Client) {
	CFdot.logger = logger
	CFdot.output = output
	CFdot.bbsClient = bbsClient
}
