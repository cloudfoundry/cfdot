package commands

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

// errors
var (
	errMissingCACertFile            = errors.New("--caCertFile must be specified if using HTTPS and --skipCertVerify is not set")
	errMissingClientCertAndKeyFiles = errors.New("--clientCertFile and --clientKeyFile must both be specified for TLS connections.")
)

func AddTLSFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&clientConfig.SkipCertVerify, "skipCertVerify", false, "when set to true, skips all SSL/TLS certificate verification [environment variable equivalent: SKIP_CERT_VERIFY]")
	cmd.Flags().StringVar(&clientConfig.BBSUrl, "bbsURL", "", "URL of BBS server to target [environment variable equivalent: BBS_URL]")
	cmd.Flags().StringVar(&clientConfig.CACertFile, "caCertFile", "", "path the Certificate Authority (CA) file to use when verifying TLS keypairs [environment variable equivalent: CA_CERT_FILE]")
	cmd.Flags().StringVar(&clientConfig.CertFile, "clientCertFile", "", "path to the TLS client certificate to use during mutual-auth TLS [environment variable equivalent: CLIENT_CERT_FILE]")
	cmd.Flags().StringVar(&clientConfig.KeyFile, "clientKeyFile", "", "path to the TLS client private key file to use during mutual-auth TLS [environment variable equivalent: CLIENT_KEY_FILE]")
	cmd.PreRunE = ClientPrehook
}

func ClientPrehook(cmd *cobra.Command, args []string) error {
	var err, returnErr error

	if clientConfig.BBSUrl == "" {
		clientConfig.BBSUrl = os.Getenv("BBS_URL")
	}

	// Only look at the environment variable if the flag has not been set.
	if !cmd.Flags().Lookup("skipCertVerify").Changed && os.Getenv("SKIP_CERT_VERIFY") != "" {
		clientConfig.SkipCertVerify, err = strconv.ParseBool(os.Getenv("SKIP_CERT_VERIFY"))
		if err != nil {
			returnErr = NewCFDotValidationError(
				cmd,
				fmt.Errorf(
					"The value '%s' is not a valid value for SKIP_CERT_VERIFY. Please specify one of the following valid boolean values: 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False",
					os.Getenv("SKIP_CERT_VERIFY")),
			)
			return returnErr
		}
	}

	if clientConfig.CACertFile == "" {
		clientConfig.CACertFile = os.Getenv("CA_CERT_FILE")
	}

	if clientConfig.CertFile == "" {
		clientConfig.CertFile = os.Getenv("CLIENT_CERT_FILE")
	}

	if clientConfig.KeyFile == "" {
		clientConfig.KeyFile = os.Getenv("CLIENT_KEY_FILE")
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
				returnErr = NewCFDotValidationError(cmd, errMissingCACertFile)
				return returnErr
			}

			err := validateReadableFile(cmd, clientConfig.CACertFile, "CA cert")
			if err != nil {
				return err
			}
		}

		if (clientConfig.KeyFile == "") || (clientConfig.CertFile == "") {
			returnErr = NewCFDotValidationError(cmd, errMissingClientCertAndKeyFiles)
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
