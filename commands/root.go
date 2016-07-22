package commands

import (
	"fmt"
	"io"
	"os"

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

func reportErr(stderr io.Writer, err error) {
	fmt.Fprintf(stderr, "error: %s\n", err.Error())
	os.Exit(1)
}
