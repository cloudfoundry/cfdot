package commands

import (
	"fmt"
	"os"
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
	cmd.Flags().StringVarP(&ttl, "ttl", "t", "", "ttl of domain [environment variable equivalent: TTL_IN_SECONDS]")

}
func TTLAsInt() int {
	return ttlAsInt
}

func SetDomainPrehook(cmd *cobra.Command, args []string) error {
	var err error

	if ttl == "" {
		ttl = os.Getenv("TTL_IN_SECONDS")
	}

	if ttl == "" {
		ttl = "0"
	}

	ttlAsInt, err = strconv.Atoi(ttl)

	if err != nil {
		return CFDotError{
			fmt.Sprintf("ttl is non-numeric"),
			3,
		}
	}

	if ttlAsInt < 0 {
		return CFDotError{
			fmt.Sprintf("ttl is negative"),
			3,
		}
	}

	if len(args) == 0 || args[0] == "" {
		return CFDotError{"No domain given", 3}
	}

	return nil
}
