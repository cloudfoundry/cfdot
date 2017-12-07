package commands

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"code.cloudfoundry.org/cfdot/commands/helpers"
	"github.com/spf13/cobra"
)

var (
	locketClientConfig helpers.TLSConfig
	locketPreHooks     = []func(cmd *cobra.Command, args []string) error{}
)

// errors
var (
	errMissingLocketUrl = errors.New("Locket API Location not set. Please specify one with the '--locketAPILocation' flag or the 'LOCKET_API_LOCATION' environment variable.")
)

func AddLocketFlags(cmd *cobra.Command) {
	AddTLSFlags(cmd)
	cmd.Flags().StringVar(&locketClientConfig.LocketApiLocation, "locketAPILocation", "", "Hostname:Port of Locket server to target [environment variable equivalent: LOCKET_API_LOCATION]")
	cmd.Flags().BoolVar(&locketClientConfig.SkipCertVerify, "locketSkipCertVerify", false, "when set to true, skips all SSL/TLS certificate verification [environment variable equivalent: LOCKET_SKIP_CERT_VERIFY]. Deprecated in favor of --skipCertVerify.")
	cmd.Flags().StringVar(&locketClientConfig.CertFile, "locketCertFile", "", "path to the TLS client certificate to use during mutual-auth TLS [environment variable equivalent: LOCKET_CERT_FILE]. Deprecated in favor of --clientCertFile.")
	cmd.Flags().StringVar(&locketClientConfig.KeyFile, "locketKeyFile", "", "path to the TLS client private key file to use during mutual-auth TLS [environment variable equivalent: LOCKET_KEY_FILE]. Deprecated in favor of --clientKeyFile.")
	cmd.Flags().StringVar(&locketClientConfig.CACertFile, "locketCACertFile", "", "path the Certificate Authority (CA) file to use when verifying TLS keypairs [environment variable equivalent: LOCKET_CA_CERT_FILE]. Deprecated in favor of --caCertFile.")
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

	locketClientConfig.Merge(Config)
	err = setLocketFlags(cmd, args)
	if err != nil {
		return err
	}

	Config = locketClientConfig
	return nil
}

func setLocketFlags(cmd *cobra.Command, args []string) error {
	var err, returnErr error
	if locketClientConfig.LocketApiLocation == "" {
		locketClientConfig.LocketApiLocation = os.Getenv("LOCKET_API_LOCATION")
	}

	// Only look at the environment variable if the flag has not been set.
	if !cmd.Flags().Lookup("locketSkipCertVerify").Changed && os.Getenv("LOCKET_SKIP_CERT_VERIFY") != "" {
		locketClientConfig.SkipCertVerify, err = strconv.ParseBool(os.Getenv("LOCKET_SKIP_CERT_VERIFY"))
		if err != nil {
			returnErr = NewCFDotValidationError(
				cmd,
				fmt.Errorf(
					"The value '%s' is not a valid value for LOCKET_SKIP_CERT_VERIFY. Please specify one of the following valid boolean values: 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False",
					os.Getenv("LOCKET_SKIP_CERT_VERIFY")),
			)
			return returnErr
		}
	}

	if locketClientConfig.CertFile == "" {
		locketClientConfig.CertFile = os.Getenv("LOCKET_CERT_FILE")
	}

	if locketClientConfig.KeyFile == "" {
		locketClientConfig.KeyFile = os.Getenv("LOCKET_KEY_FILE")
	}

	if locketClientConfig.CACertFile == "" {
		locketClientConfig.CACertFile = os.Getenv("LOCKET_CA_CERT_FILE")
	}

	if locketClientConfig.LocketApiLocation == "" {
		returnErr = NewCFDotValidationError(cmd, errMissingLocketUrl)
		return returnErr
	}

	if !locketClientConfig.SkipCertVerify {
		if locketClientConfig.CACertFile == "" {
			returnErr = NewCFDotValidationError(cmd, errMissingCACertFile)
			return returnErr
		}

		err := validateReadableFile(cmd, locketClientConfig.CACertFile, "CA cert")

		if err != nil {
			return err
		}
	}

	if (locketClientConfig.KeyFile == "") || (locketClientConfig.CertFile == "") {
		returnErr = NewCFDotValidationError(cmd, errMissingClientCertAndKeyFiles)
		return returnErr
	}

	if locketClientConfig.KeyFile != "" {
		err := validateReadableFile(cmd, locketClientConfig.KeyFile, "key")

		if err != nil {
			return err
		}
	}

	if locketClientConfig.CertFile != "" {
		err := validateReadableFile(cmd, locketClientConfig.CertFile, "cert")

		if err != nil {
			return err
		}
	}

	return nil
}
