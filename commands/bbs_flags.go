package commands

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"net/url"

	"code.cloudfoundry.org/bbs"
	"github.com/spf13/cobra"
)

const (
	clientSessionCacheSize int = 0
	maxIdleConnsPerHost    int = 0
)

// flags
var (
	bbsURL            string
	bbsCACertFile     string
	bbsCertFile       string
	bbsKeyFile        string
	bbsSkipCertVerify bool
)

// errors
var (
	errMissingBBSUrl          = errors.New("BBS URL not set. Please specify one with the '--bbsURL' flag or the 'BBS_URL' environment variable.")
	errMissingCACertFile      = errors.New("--bbsCACertFile must be specified if using HTTPS and --bbsSkipCertVerify is not set")
	errMissingCertAndKeyFiles = errors.New("--bbsCertFile and --bbsKeyFile must both be specified for TLS connections.")
)

func AddBBSFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&bbsSkipCertVerify, "bbsSkipCertVerify", false, "when set to true, skips all SSL/TLS certificate verification [environment variable equivalent: BBS_SKIP_CERT_VERIFY]")
	cmd.Flags().StringVar(&bbsURL, "bbsURL", "", "URL of BBS server to target [environment variable equivalent: BBS_URL]")
	cmd.Flags().StringVar(&bbsCertFile, "bbsCertFile", "", "path to the TLS client certificate to use during mutual-auth TLS [environment variable equivalent: BBS_CERT_FILE]")
	cmd.Flags().StringVar(&bbsKeyFile, "bbsKeyFile", "", "path to the TLS client private key file to use during mutual-auth TLS [environment variable equivalent: BBS_KEY_FILE]")
	cmd.Flags().StringVar(&bbsCACertFile, "bbsCACertFile", "", "path the Certificate Authority (CA) file to use when verifying TLS keypairs [environment variable equivalent: BBS_CA_CERT_FILE]")
	cmd.PreRunE = BBSPrehook
}

func BBSPrehook(cmd *cobra.Command, args []string) error {
	var err, returnErr error

	if bbsURL == "" {
		bbsURL = os.Getenv("BBS_URL")
	}

	// Only look at the environment variable if the flag has not been set.
	if !cmd.Flags().Lookup("bbsSkipCertVerify").Changed && os.Getenv("BBS_SKIP_CERT_VERIFY") != "" {
		bbsSkipCertVerify, err = strconv.ParseBool(os.Getenv("BBS_SKIP_CERT_VERIFY"))
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

	if bbsCertFile == "" {
		bbsCertFile = os.Getenv("BBS_CERT_FILE")
	}

	if bbsKeyFile == "" {
		bbsKeyFile = os.Getenv("BBS_KEY_FILE")
	}

	if bbsCACertFile == "" {
		bbsCACertFile = os.Getenv("BBS_CA_CERT_FILE")
	}

	if bbsURL == "" {
		returnErr = NewCFDotValidationError(cmd, errMissingBBSUrl)
		return returnErr
	}

	var parsedURL *url.URL
	if parsedURL, err = url.Parse(bbsURL); err != nil {
		returnErr = NewCFDotValidationError(
			cmd,
			fmt.Errorf(
				"The value '%s' is not a valid BBS URL. Please specify one with the '--bbsURL' flag or the 'BBS_URL' environment variable.",
				bbsURL),
		)
		return returnErr
	}

	if parsedURL.Scheme == "https" {
		if !bbsSkipCertVerify {
			if bbsCACertFile == "" {
				returnErr = NewCFDotValidationError(cmd, errMissingCACertFile)
				return returnErr
			}

			err := validateReadableFile(bbsCACertFile)

			if err != nil {
				returnErr = NewCFDotValidationError(
					cmd,
					fmt.Errorf("CA cert file '%s' doesn't exist or is not readable: %s", bbsCACertFile, err.Error()),
				)
				return returnErr
			}
		}

		if (bbsKeyFile == "") || (bbsCertFile == "") {
			returnErr = NewCFDotValidationError(cmd, errMissingCertAndKeyFiles)
			return returnErr
		}

		if bbsKeyFile != "" {
			err := validateReadableFile(bbsKeyFile)

			if err != nil {
				returnErr = NewCFDotValidationError(
					cmd,
					fmt.Errorf("key file '%s' doesn't exist or is not readable: %s", bbsKeyFile, err.Error()),
				)
				return returnErr
			}
		}

		if bbsCertFile != "" {
			err := validateReadableFile(bbsCertFile)

			if err != nil {
				returnErr = NewCFDotValidationError(
					cmd,
					fmt.Errorf("cert file '%s' doesn't exist or is not readable: %s", bbsCertFile, err.Error()),
				)
				return returnErr
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
					"variable.", bbsURL),
		)
		return returnErr
	}

	return nil
}

func newBBSClient(cmd *cobra.Command) (bbs.Client, error) {
	var err error
	var client bbs.Client

	if !strings.HasPrefix(bbsURL, "https") {
		client = bbs.NewClient(bbsURL)
	} else {
		if bbsSkipCertVerify {
			client, err = bbs.NewSecureSkipVerifyClient(
				bbsURL,
				bbsCertFile,
				bbsKeyFile,
				clientSessionCacheSize,
				maxIdleConnsPerHost,
			)
		} else {
			client, err = bbs.NewSecureClient(
				bbsURL,
				bbsCACertFile,
				bbsCertFile,
				bbsKeyFile,
				clientSessionCacheSize,
				maxIdleConnsPerHost,
			)
		}
	}

	return client, err
}

func validateReadableFile(filename string) error {
	file, err := os.Open(filename)
	defer file.Close()

	return err
}
