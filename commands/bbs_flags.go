package commands

import (
	"fmt"
	"io/ioutil"
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
	var bbsSkipCertVerifyAsString string
	cmd.Flags().StringVar(&bbsURL, "bbsURL", "", "URL of BBS server to target, can also be specified with BBS_URL environment variable")
	// Read this in as a StringVar instead of a BoolVar so we can check whether it was set or not, and use an environment variable if it was not set.
	cmd.Flags().StringVar(&bbsSkipCertVerifyAsString, "bbsSkipCertVerify", "", "when set to true, skips all SSL/TLS certificate verification")
	cmd.Flags().StringVar(&bbsCertFile, "bbsCertFile", "", "path to the TLS client certificate to use during mutual-auth TLS")
	cmd.Flags().StringVar(&bbsKeyFile, "bbsKeyFile", "", "path to the TLS client private key file to use during mutual-auth TLS")
	cmd.Flags().StringVar(&bbsCACertFile, "bbsCACertFile", "", "path the Certificate Authority (CA) file to use when verifying TLS keypairs")

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		var returnErr error

		if bbsURL == "" {
			bbsURL = os.Getenv("BBS_URL")
		}

		if bbsSkipCertVerifyAsString == "" {
			bbsSkipCertVerifyAsString = os.Getenv("BBS_SKIP_CERT_VERIFY")
			if bbsSkipCertVerifyAsString == "" {
				bbsSkipCertVerifyAsString = "false"
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

		var err error
		bbsSkipCertVerify, err = strconv.ParseBool(bbsSkipCertVerifyAsString)
		if err != nil {
			returnErr = CFDotError{
				fmt.Sprintf("The value '%s' is not a valid value for bbsSkipCertVerify. Please specify one of the following valid boolean values: 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False", bbsSkipCertVerifyAsString),
				3}
			return returnErr
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

				_, err := ioutil.ReadFile(bbsCACertFile)
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
				_, err := ioutil.ReadFile(bbsKeyFile)
				if err != nil {
					returnErr = CFDotError{
						fmt.Sprintf("key file '"+bbsKeyFile+"' doesn't exist or is not readable: %s", err.Error()),
						3}
					return returnErr
				}
			}

			if bbsCertFile != "" {
				_, err := ioutil.ReadFile(bbsCertFile)
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
