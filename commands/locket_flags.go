package commands

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket"
	locketmodels "code.cloudfoundry.org/locket/models"
	"github.com/spf13/cobra"
)

// flags
var (
	locketAPILocation    string
	locketCACertFile     string
	locketCertFile       string
	locketKeyFile        string
	locketSkipCertVerify bool
)

// errors
var (
	errMissingLocketUrl             = errors.New("Locket API Location not set. Please specify one with the '--locketAPILocation' flag or the 'LOCKET_API_LOCATION' environment variable.")
	errMissingLocketCACertFile      = errors.New("--locketCACertFile must be specified if --locketSkipCertVerify is not set")
	errMissingLocketCertAndKeyFiles = errors.New("--locketCertFile and --locketKeyFile must both be specified for TLS connections.")
)

func AddLocketFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&locketSkipCertVerify, "locketSkipCertVerify", false, "when set to true, skips all SSL/TLS certificate verification [environment variable equivalent: LOCKET_SKIP_CERT_VERIFY]")
	cmd.Flags().StringVar(&locketAPILocation, "locketAPILocation", "", "Hostname:Port of Locket server to target [environment variable equivalent: LOCKET_API_LOCATION]")
	cmd.Flags().StringVar(&locketCertFile, "locketCertFile", "", "path to the TLS client certificate to use during mutual-auth TLS [environment variable equivalent: LOCKET_CERT_FILE]")
	cmd.Flags().StringVar(&locketKeyFile, "locketKeyFile", "", "path to the TLS client private key file to use during mutual-auth TLS [environment variable equivalent: LOCKET_KEY_FILE]")
	cmd.Flags().StringVar(&locketCACertFile, "locketCACertFile", "", "path the Certificate Authority (CA) file to use when verifying TLS keypairs [environment variable equivalent: LOCKET_CA_CERT_FILE]")
	cmd.PreRunE = LocketPrehook
}

func LocketPrehook(cmd *cobra.Command, args []string) error {
	var err, returnErr error
	if locketAPILocation == "" {
		locketAPILocation = os.Getenv("LOCKET_API_LOCATION")
	}

	// Only look at the environment variable if the flag has not been set.
	if !cmd.Flags().Lookup("locketSkipCertVerify").Changed && os.Getenv("LOCKET_SKIP_CERT_VERIFY") != "" {
		locketSkipCertVerify, err = strconv.ParseBool(os.Getenv("LOCKET_SKIP_CERT_VERIFY"))
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

	if locketCertFile == "" {
		locketCertFile = os.Getenv("LOCKET_CERT_FILE")
	}

	if locketKeyFile == "" {
		locketKeyFile = os.Getenv("LOCKET_KEY_FILE")
	}

	if locketCACertFile == "" {
		locketCACertFile = os.Getenv("LOCKET_CA_CERT_FILE")
	}

	if locketAPILocation == "" {
		returnErr = NewCFDotValidationError(cmd, errMissingLocketUrl)
		return returnErr
	}

	if !locketSkipCertVerify {
		if locketCACertFile == "" {
			returnErr = NewCFDotValidationError(cmd, errMissingLocketCACertFile)
			return returnErr
		}

		err := validateReadableFile(cmd, locketCACertFile, "CA cert")

		if err != nil {
			return err
		}
	}

	if (locketKeyFile == "") || (locketCertFile == "") {
		returnErr = NewCFDotValidationError(cmd, errMissingLocketCertAndKeyFiles)
		return returnErr
	}

	if locketKeyFile != "" {
		err := validateReadableFile(cmd, locketKeyFile, "key")

		if err != nil {
			return err
		}
	}

	if locketCertFile != "" {
		err := validateReadableFile(cmd, locketCertFile, "cert")

		if err != nil {
			return err
		}
	}

	return nil
}

func newLocketClient(logger lager.Logger, cmd *cobra.Command) (locketmodels.LocketClient, error) {
	var err error
	var client locketmodels.LocketClient
	config := locket.ClientLocketConfig{
		LocketAddress:        locketAPILocation,
		LocketCACertFile:     locketCACertFile,
		LocketClientCertFile: locketCertFile,
		LocketClientKeyFile:  locketKeyFile,
	}

	if locketSkipCertVerify {
		client, err = locket.NewClientSkipCertVerify(
			logger,
			config,
		)
	} else {
		client, err = locket.NewClient(
			logger,
			config,
		)
	}

	return client, err
}
