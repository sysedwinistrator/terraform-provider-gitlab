package utils

import (
	"fmt"
	"os"
	"strconv"
)

// ParseConfigBoolFromEnv parses the given environment variable as boolean
func ParseConfigBoolFromEnv(varName string, defaultValue bool) (bool, error) {
	v := os.Getenv(varName)
	if v == "" {
		return defaultValue, nil
	}

	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("The environment variable '%s' with value '%s' cannot be parsed as bool: %s", varName, v, err)
	}

	return b, nil
}
