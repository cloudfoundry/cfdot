package commands

import (
	"errors"
	"fmt"
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
