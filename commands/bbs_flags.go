package commands

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"code.cloudfoundry.org/cfdot/commands/helpers"

	"net/url"

	"github.com/spf13/cobra"
)

// flags
var (
	clientConfig helpers.ClientConfig
)

// errors
var (
	errMissingBBSUrl             = errors.New("BBS URL not set. Please specify one with the '--bbsURL' flag or the 'BBS_URL' environment variable.")
	errMissingBBSCACertFile      = errors.New("--bbsCACertFile must be specified if using HTTPS and --bbsSkipCertVerify is not set")
	errMissingBBSCertAndKeyFiles = errors.New("--bbsCertFile and --bbsKeyFile must both be specified for TLS connections.")
)

func AddBBSFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&clientConfig.SkipCertVerify, "bbsSkipCertVerify", false, "when set to true, skips all SSL/TLS certificate verification [environment variable equivalent: BBS_SKIP_CERT_VERIFY]")
	cmd.Flags().StringVar(&clientConfig.BBSUrl, "bbsURL", "", "URL of BBS server to target [environment variable equivalent: BBS_URL]")
	cmd.Flags().StringVar(&clientConfig.CertFile, "bbsCertFile", "", "path to the TLS client certificate to use during mutual-auth TLS [environment variable equivalent: BBS_CERT_FILE]")
	cmd.Flags().StringVar(&clientConfig.KeyFile, "bbsKeyFile", "", "path to the TLS client private key file to use during mutual-auth TLS [environment variable equivalent: BBS_KEY_FILE]")
	cmd.Flags().StringVar(&clientConfig.CACertFile, "bbsCACertFile", "", "path the Certificate Authority (CA) file to use when verifying TLS keypairs [environment variable equivalent: BBS_CA_CERT_FILE]")
	cmd.PreRunE = BBSPrehook
}

func BBSPrehook(cmd *cobra.Command, args []string) error {
	var err, returnErr error

	if clientConfig.BBSUrl == "" {
		clientConfig.BBSUrl = os.Getenv("BBS_URL")
	}

	// Only look at the environment variable if the flag has not been set.
	if !cmd.Flags().Lookup("bbsSkipCertVerify").Changed && os.Getenv("BBS_SKIP_CERT_VERIFY") != "" {
		clientConfig.SkipCertVerify, err = strconv.ParseBool(os.Getenv("BBS_SKIP_CERT_VERIFY"))
		if err != nil {
			returnErr = NewCFDotValidationError(
				cmd,
				fmt.Errorf(
					"The value '%s' is not a valid value for BBS_SKIP_CERT_VERIFY. Please specify one of the following valid boolean values: 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False",
					os.Getenv("BBS_SKIP_CERT_VERIFY")),
			)
			return returnErr
		}
	}

	if clientConfig.CertFile == "" {
		clientConfig.CertFile = os.Getenv("BBS_CERT_FILE")
	}

	if clientConfig.KeyFile == "" {
		clientConfig.KeyFile = os.Getenv("BBS_KEY_FILE")
	}

	if clientConfig.CACertFile == "" {
		clientConfig.CACertFile = os.Getenv("BBS_CA_CERT_FILE")
	}

	if clientConfig.BBSUrl == "" {
		returnErr = NewCFDotValidationError(cmd, errMissingBBSUrl)
		return returnErr
	}

	var parsedURL *url.URL
	if parsedURL, err = url.Parse(clientConfig.BBSUrl); err != nil {
		returnErr = NewCFDotValidationError(
			cmd,
			fmt.Errorf(
				"The value '%s' is not a valid BBS URL. Please specify one with the '--bbsURL' flag or the 'BBS_URL' environment variable.",
				clientConfig.BBSUrl),
		)
		return returnErr
	}

	if parsedURL.Scheme == "https" {
		if !clientConfig.SkipCertVerify {
			if clientConfig.CACertFile == "" {
				returnErr = NewCFDotValidationError(cmd, errMissingBBSCACertFile)
				return returnErr
			}

			err := validateReadableFile(cmd, clientConfig.CACertFile, "CA cert")
			if err != nil {
				return err
			}
		}

		if (clientConfig.KeyFile == "") || (clientConfig.CertFile == "") {
			returnErr = NewCFDotValidationError(cmd, errMissingBBSCertAndKeyFiles)
			return returnErr
		}

		if clientConfig.KeyFile != "" {
			err := validateReadableFile(cmd, clientConfig.KeyFile, "key")

			if err != nil {
				return err
			}
		}

		if clientConfig.CertFile != "" {
			err := validateReadableFile(cmd, clientConfig.CertFile, "cert")

			if err != nil {
				return err
			}
		}

		return nil
	}

	if parsedURL.Scheme != "http" {
		returnErr = NewCFDotValidationError(
			cmd,
			fmt.Errorf(
				"The URL '%s' does not have an 'http' or 'https' scheme. Please "+
					"specify one with the '--bbsURL' flag or the 'BBS_URL' environment "+
					"variable.", clientConfig.BBSUrl),
		)
		return returnErr
	}

	return nil
}

func validateReadableFile(cmd *cobra.Command, filename, filetype string) error {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return NewCFDotValidationError(
			cmd,
			fmt.Errorf("%s file '%s' doesn't exist or is not readable: %s", filetype, filename, err.Error()),
		)
	}
	return nil
}
