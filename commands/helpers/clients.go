package helpers

import (
	"strings"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/rep"
	"github.com/spf13/cobra"
)

const (
	clientSessionCacheSize int = -1
	maxIdleConnsPerHost    int = -1
)

type ClientConfig struct {
	BBSUrl         string
	CACertFile     string
	CertFile       string
	KeyFile        string
	SkipCertVerify bool
}

func NewBBSClient(cmd *cobra.Command, clientConfig ClientConfig) (bbs.Client, error) {
	var err error
	var client bbs.Client

	if !strings.HasPrefix(clientConfig.BBSUrl, "https") {
		client = bbs.NewClient(clientConfig.BBSUrl)
	} else {
		if clientConfig.SkipCertVerify {
			client, err = bbs.NewSecureSkipVerifyClient(
				clientConfig.BBSUrl,
				clientConfig.CertFile,
				clientConfig.KeyFile,
				clientSessionCacheSize,
				maxIdleConnsPerHost,
			)
		} else {
			client, err = bbs.NewSecureClient(
				clientConfig.BBSUrl,
				clientConfig.CACertFile,
				clientConfig.CertFile,
				clientConfig.KeyFile,
				clientSessionCacheSize,
				maxIdleConnsPerHost,
			)
		}
	}

	return client, err
}

func NewRepClient(clientFactory rep.ClientFactory, address, url string) (rep.Client, error) {
	return clientFactory.CreateClient(address, url)
}
