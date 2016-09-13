package commands

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

func ValidatePositiveIntegerForFlag(flag, value string, cmd *cobra.Command) (int, error) {
	errorInvalidInteger := errors.New(fmt.Sprintf("%s is non-numeric", flag))
	errorNegativeInteger := errors.New(fmt.Sprintf("%s is negative", flag))

	valueAsInt, err := strconv.Atoi(value)

	if err != nil {
		return -1, NewCFDotValidationError(cmd, errorInvalidInteger)
	}

	if valueAsInt < 0 {
		return -1, CFDotError{
			errorNegativeInteger,
			3,
		}
	}

	return valueAsInt, nil
}

func ValidateConflictingShortAndLongFlag(short string, long string, cmd *cobra.Command) error {
	errorConflictingShortAndLongFlagPassed := errors.New(fmt.Sprintf("Only one of %s and %s should be passed", short, long))

	if contains(os.Args, short) && contains(os.Args, long) {
		return NewCFDotValidationError(cmd, errorConflictingShortAndLongFlagPassed)
	}

	return nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}
