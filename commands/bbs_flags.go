package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"net/url"

	"code.cloudfoundry.org/bbs"
	"github.com/spf13/cobra"
)

var (
	bbsURL            string
	bbsCACertFile     string
	bbsCertFile       string
	bbsKeyFile        string
	bbsSkipCertVerify bool
)

const clientSessionCacheSize int = 0
const maxIdleConnsPerHost int = 0

func AddBBSFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&bbsURL, "bbsURL", "", "URL of BBS server to target [environment variable equivalent: BBS_URL]")
	// Read this in as a StringVar instead of a BoolVar so we can check whether it was set or not, and use an environment variable if it was not set.
	cmd.Flags().BoolVar(&bbsSkipCertVerify, "bbsSkipCertVerify", false, "when set to true, skips all SSL/TLS certificate verification [environment variable equivalent: BBS_SKIP_CERT_VERIFY]")
	cmd.Flags().StringVar(&bbsCertFile, "bbsCertFile", "", "path to the TLS client certificate to use during mutual-auth TLS [environment variable equivalent: BBS_CERT_FILE]")
	cmd.Flags().StringVar(&bbsKeyFile, "bbsKeyFile", "", "path to the TLS client private key file to use during mutual-auth TLS [environment variable equivalent: BBS_KEY_FILE]")
	cmd.Flags().StringVar(&bbsCACertFile, "bbsCACertFile", "", "path the Certificate Authority (CA) file to use when verifying TLS keypairs [environment variable equivalent: BBS_CA_CERT_FILE]")

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		var err, returnErr error

		if bbsURL == "" {
			bbsURL = os.Getenv("BBS_URL")
		}

		// Only look at the environment variable if the flag has not been set.
		if !cmd.Flags().Lookup("bbsSkipCertVerify").Changed && os.Getenv("BBS_SKIP_CERT_VERIFY") != "" {
			bbsSkipCertVerify, err = strconv.ParseBool(os.Getenv("BBS_SKIP_CERT_VERIFY"))
			if err != nil {
				returnErr = CFDotError{
					fmt.Sprintf(
						"The value '%s' is not a valid value for BBS_SKIP_CERT_VERIFY. Please specify one of the following valid boolean values: 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False",
						os.Getenv("BBS_SKIP_CERT_VERIFY")),
					3}
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
			returnErr = CFDotError{
				"BBS URL not set. Please specify one with the '--bbsURL' flag or the 'BBS_URL' environment variable.",
				3}
			return returnErr
		}

		var parsedURL *url.URL
		if parsedURL, err = url.Parse(bbsURL); err != nil {
			returnErr = CFDotError{
				fmt.Sprintf(
					"The value '%s' is not a valid BBS URL. Please specify one with the '--bbsURL' flag or the 'BBS_URL' environment variable.",
					bbsURL),
				3}
			return returnErr
		}

		if parsedURL.Scheme == "https" {
			if !bbsSkipCertVerify {
				if bbsCACertFile == "" {
					returnErr = CFDotError{"--bbsCACertFile must be specified if using HTTPS and --bbsSkipCertVerify is not set", 3}
					return returnErr
				}

				err := validateReadableFile(bbsCACertFile)

				if err != nil {
					returnErr = CFDotError{
						fmt.Sprintf("CA cert file '"+bbsCACertFile+"' doesn't exist or is not readable: %s", err.Error()),
						3}
					return returnErr
				}
			}

			if (bbsKeyFile == "") != (bbsCertFile == "") {
				returnErr = CFDotError{
					"--bbsCertFile and --bbsKeyFile must both be specified for mutual TLS connections",
					3}
				return returnErr
			}

			if bbsKeyFile != "" {
				err := validateReadableFile(bbsKeyFile)

				if err != nil {
					returnErr = CFDotError{
						fmt.Sprintf("key file '"+bbsKeyFile+"' doesn't exist or is not readable: %s", err.Error()),
						3}
					return returnErr
				}
			}

			if bbsCertFile != "" {
				err := validateReadableFile(bbsCertFile)

				if err != nil {
					returnErr = CFDotError{
						fmt.Sprintf("cert file '"+bbsCertFile+"' doesn't exist or is not readable: %s", err.Error()),
						3}
					return returnErr
				}
			}

			return nil
		}

		if parsedURL.Scheme != "http" {
			returnErr = CFDotError{
				fmt.Sprintf(
					"The URL '%s' does not have an 'http' or 'https' scheme. Please "+
						"specify one with the '--bbsURL' flag or the 'BBS_URL' environment "+
						"variable.", bbsURL),
				3}
			return returnErr
		}

		return nil
	}
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
