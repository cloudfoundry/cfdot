package commands

import (
	"errors"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	ttl      string
	ttlAsInt int
)

func AddSetDomainFlags(cmd *cobra.Command) {

	// Read this in as a StringVar so we can check whether it was set or not, and
	// use an environment variable if not set, and throw our own error instead of
	// using the error from pflag
	cmd.Flags().StringVarP(&ttl, "ttl", "t", "", "ttl of domain")

}
func TTLAsInt() int {
	return ttlAsInt
}

var (
	errMissingDomain = errors.New("No domain given")
	errInvalidTTL    = errors.New("ttl is non-numeric")
	errNegativeTTL   = errors.New("ttl is negative")
)

func SetDomainPrehook(cmd *cobra.Command, args []string) error {
	var err error

	if ttl == "" {
		ttl = "0"
	}

	ttlAsInt, err = strconv.Atoi(ttl)

	if err != nil {
		return NewCFDotValidationError(cmd, errInvalidTTL)
	}

	if ttlAsInt < 0 {
		return CFDotError{
			errNegativeTTL,
			3,
		}
	}

	if len(args) == 0 || args[0] == "" {
		return NewCFDotValidationError(cmd, errMissingDomain)
	}

	return nil
}
