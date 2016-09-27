package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/bbs/models"
	"github.com/spf13/cobra"
)

var createDesiredLRPCmd = &cobra.Command{
	Use:   "create-desired-lrp (spec|@file)",
	Short: "Create a desired LRP",
	Long:  "Create a desired LRP from the given spec. Spec can either be json encoded desired-lrp, e.g. '{\"process_guid\":\"some-guid\"}' or a file containing json encoded desired-lrp, e.g. @/path/to/spec/file",
	RunE:  createDesiredLRP,
}

func init() {
	AddBBSFlags(createDesiredLRPCmd)
	RootCmd.AddCommand(createDesiredLRPCmd)
}

func createDesiredLRP(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return NewCFDotValidationError(cmd, fmt.Errorf("missing spec argument"))
	}

	spec, err := ValidateCreateDesiredLRPArguments(args)
	if err != nil {
		return NewCFDotValidationError(cmd, err)
	}

	bbsClient, err := newBBSClient(cmd)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	err = CreateDesiredLRP(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, spec)
	if err != nil {
		return NewCFDotError(cmd, err)
	}

	return nil
}

func ValidateCreateDesiredLRPArguments(args []string) ([]byte, error) {
	var desiredLRP *models.DesiredLRP
	var err error
	var spec []byte
	argValue := args[0]
	if strings.HasPrefix(argValue, "@") {
		_, err := os.Stat(argValue[1:])
		if err != nil {
			return nil, err
		}
		spec, err = ioutil.ReadFile(argValue[1:])
		if err != nil {
			return nil, err
		}

	} else {
		spec = []byte(argValue)
	}
	err = json.Unmarshal([]byte(spec), &desiredLRP)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Invalid JSON: %s", err.Error()))
	}
	return spec, nil
}

func CreateDesiredLRP(stdout, stderr io.Writer, bbsClient bbs.Client, spec []byte) error {
	logger := globalLogger.Session("create-desired-lrp")

	var desiredLRP *models.DesiredLRP
	err := json.Unmarshal(spec, &desiredLRP)
	if err != nil {
		return err
	}
	err = bbsClient.DesireLRP(logger, desiredLRP)
	if err != nil {
		return err
	}

	return nil
}
