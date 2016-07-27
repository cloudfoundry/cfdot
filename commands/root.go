package commands

import (
	"errors"
	"fmt"
	"os"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/lager"
	"github.com/spf13/cobra"
	"net/url"
)

var logger = lager.NewLogger("cfdot")

var RootCmd = &cobra.Command{
	Use:   "cfdot",
	Short: "Diego operator tooling",
	Long:  "A command-line tool to interact with a Cloud Foundry Diego deployment",
}

var bbsURL string

func addBBSFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&bbsURL, "bbsURL", "", "", "URL of BBS server to target, can also be specified with BBS_URL environment variable")
	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		bbsURLFlag := cmd.Flag("bbsURL").Value.String()

		if bbsURLFlag == "" {
			reportErr(cmd, errors.New(
				"BBS URL not set. Please specify one with the '--bbsURL' flag or the "+
					"'BBS_URL' environment variable.",
			), 3)
		} else if parsedURL, err := url.Parse(bbsURLFlag); err != nil {
			reportErr(cmd, errors.New(fmt.Sprintf(
				"The value '%s' is not a valid BBS URL. Please specify one with the "+
					"'--bbsURL' flag or the 'BBS_URL' environment variable.",
				bbsURLFlag,
			)), 3)
		} else if parsedURL.Scheme != "https" && parsedURL.Scheme != "http" {
			reportErr(cmd, errors.New(fmt.Sprintf(
				"The URL '%s' does not have an 'http' or 'https' scheme. Please "+
					"specify one with the '--bbsURL' flag or the 'BBS_URL' environment "+
					"variable.",
				bbsURLFlag,
			)), 3)
		}

	}
}

func newBBSClient(cmd *cobra.Command) bbs.Client {
	return bbs.NewClient(bbsURL)
}

func reportErr(cmd *cobra.Command, err error, exitCode int) {
	cmd.SetOutput(cmd.OutOrStderr())
	fmt.Fprintf(cmd.OutOrStderr(), "error: %s\n\n", err.Error())
	cmd.Help()
	os.Exit(exitCode)
}
