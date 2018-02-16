package commands

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
)

var (
	locketApiLocation string
	locketPreHooks    = []func(cmd *cobra.Command, args []string) error{}
)

// errors
var (
	errMissingLocketUrl = errors.New("Locket API Location not set. Please specify one with the '--locketAPILocation' flag or the 'LOCKET_API_LOCATION' environment variable.")
)

func AddLocketFlags(cmd *cobra.Command) {
	AddTLSFlags(cmd)
	cmd.Flags().StringVar(&locketApiLocation, "locketAPILocation", "", "Hostname:Port of Locket server to target [environment variable equivalent: LOCKET_API_LOCATION]")
	locketPreHooks = append(locketPreHooks, cmd.PreRunE)
	cmd.PreRunE = LocketPrehook
}

func LocketPrehook(cmd *cobra.Command, args []string) error {
	var err error
	for _, f := range locketPreHooks {
		if f == nil {
			continue
		}
		err = f(cmd, args)
		if err != nil {
			return err
		}
	}

	err = setLocketFlags(cmd, args)
	if err != nil {
		return err
	}

	return nil
}

func setLocketFlags(cmd *cobra.Command, args []string) error {
	var returnErr error
	if locketApiLocation == "" {
		locketApiLocation = os.Getenv("LOCKET_API_LOCATION")
	}

	Config.LocketApiLocation = locketApiLocation
	if Config.LocketApiLocation == "" {
		returnErr = NewCFDotValidationError(cmd, errMissingLocketUrl)
		return returnErr
	}

	if !Config.SkipCertVerify {
		if Config.CACertFile == "" {
			returnErr = NewCFDotValidationError(cmd, errMissingCACertFile)
			return returnErr
		}

		err := validateReadableFile(cmd, Config.CACertFile, "CA cert")

		if err != nil {
			return err
		}
	}

	if (Config.KeyFile == "") || (Config.CertFile == "") {
		returnErr = NewCFDotValidationError(cmd, errMissingClientCertAndKeyFiles)
		return returnErr
	}

	if Config.KeyFile != "" {
		err := validateReadableFile(cmd, Config.KeyFile, "key")

		if err != nil {
			return err
		}
	}

	if Config.CertFile != "" {
		err := validateReadableFile(cmd, Config.CertFile, "cert")

		if err != nil {
			return err
		}
	}

	return nil
}
