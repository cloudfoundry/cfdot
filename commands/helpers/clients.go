package helpers

import (
	"strings"

	"code.cloudfoundry.org/bbs"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/locket"
	locketmodels "code.cloudfoundry.org/locket/models"
	"code.cloudfoundry.org/rep"
	"github.com/spf13/cobra"
)

const (
	clientSessionCacheSize int = -1
	maxIdleConnsPerHost    int = -1
)

type TLSConfig struct {
	BBSUrl            string
	LocketApiLocation string
	CACertFile        string
	CertFile          string
	KeyFile           string
	SkipCertVerify    bool
}

func NewBBSClient(cmd *cobra.Command, bbsClientConfig TLSConfig) (bbs.Client, error) {
	var err error
	var client bbs.Client

	if !strings.HasPrefix(bbsClientConfig.BBSUrl, "https") {
		client = bbs.NewClient(bbsClientConfig.BBSUrl)
	} else {
		if bbsClientConfig.SkipCertVerify {
			client, err = bbs.NewSecureSkipVerifyClient(
				bbsClientConfig.BBSUrl,
				bbsClientConfig.CertFile,
				bbsClientConfig.KeyFile,
				clientSessionCacheSize,
				maxIdleConnsPerHost,
			)
		} else {
			client, err = bbs.NewSecureClient(
				bbsClientConfig.BBSUrl,
				bbsClientConfig.CACertFile,
				bbsClientConfig.CertFile,
				bbsClientConfig.KeyFile,
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

func NewLocketClient(logger lager.Logger, cmd *cobra.Command, locketClientConfig TLSConfig) (locketmodels.LocketClient, error) {
	var err error
	var client locketmodels.LocketClient
	config := locket.ClientLocketConfig{
		LocketAddress:        locketClientConfig.LocketApiLocation,
		LocketCACertFile:     locketClientConfig.CACertFile,
		LocketClientCertFile: locketClientConfig.CertFile,
		LocketClientKeyFile:  locketClientConfig.KeyFile,
	}

	if locketClientConfig.SkipCertVerify {
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

func (config *TLSConfig) Merge(newConfig TLSConfig) {
	if newConfig.KeyFile != "" {
		config.KeyFile = newConfig.KeyFile
	}
	if newConfig.CACertFile != "" {
		config.CACertFile = newConfig.CACertFile
	}
	if newConfig.CertFile != "" {
		config.CertFile = newConfig.CertFile
	}
	config.SkipCertVerify = config.SkipCertVerify || newConfig.SkipCertVerify
}
