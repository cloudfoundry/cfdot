package commands

import (
	"errors"
	"fmt"
	"os"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/lager"
	"github.com/spf13/cobra"
)

var logger = lager.NewLogger("cfdot")

var RootCmd = &cobra.Command{
	Use:   "cfdot",
	Short: "Diego operator tooling.",
	Long:  "A command-line tool to interact with a Cloud Foundry Diego deployment.",
}

var bbsURL string

func addBBSFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&bbsURL, "bbsURL", "", "", "BBS URL")
}

func newBBSClient(cmd *cobra.Command) bbs.Client {
	if bbsURL == "" {
		reportErr(cmd, errors.New("the required flag '--bbsURL' was not specified"))
	}
	return bbs.NewClient(bbsURL)
}

func reportErr(cmd *cobra.Command, err error) {
	fmt.Fprintf(cmd.OutOrStderr(), "error: %s\n\n", err.Error())
	cmd.Help()
	os.Exit(1)
}
