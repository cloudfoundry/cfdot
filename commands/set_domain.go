package commands

import (
	"fmt"
	"io"
	"time"

	"code.cloudfoundry.org/bbs"

	"github.com/spf13/cobra"
)

func init() {
	AddBBSFlags(setDomainCmd)
	AddSetDomainFlags(setDomainCmd)
	setDomainCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		err := SetDomainPrehook(cmd, args)
		if err != nil {
			return err
		}
		return BBSPrehook(cmd, args)
	}
	RootCmd.AddCommand(setDomainCmd)
}

var setDomainCmd = &cobra.Command{
	Use:   "set-domain",
	Short: "Set domain",
	Long:  "Mark a domain as fresh for ttl seconds, where 0 or non-specified means keep fresh permanently",
	RunE:  setDomain,
}

func setDomain(cmd *cobra.Command, args []string) error {
	var err error
	var bbsClient bbs.Client

	bbsClient, err = newBBSClient(cmd)
	if err != nil {
		return CFDotError{err.Error(), 1}
	}

	err = SetDomain(cmd.OutOrStdout(), cmd.OutOrStderr(), bbsClient, args, ttlAsInt)
	if err != nil {
		return CFDotError{err.Error(), 1}
	}

	return nil
}

func SetDomain(stdout, stderr io.Writer, bbsClient bbs.Client, args []string, ttl int) error {
	logger := globalLogger.Session("set-domain")

	var duration time.Duration = time.Duration(ttl) * time.Second

	// The prehook catches the case where we don't specify any args
	domain := args[0]

	err := bbsClient.UpsertDomain(logger, domain, duration)
	if err != nil {
		return err
	}

	io.WriteString(
		stdout,
		fmt.Sprintf("Set domain \"%s\" with TTL at %d seconds", domain, ttl),
	)

	return nil
}
