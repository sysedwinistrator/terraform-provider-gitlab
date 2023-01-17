package utils

import (
	"fmt"
	"strings"
)

// return the pieces of id `a:b` as a, b
func ParseTwoPartID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected ID format (%q). Expected project:key", id)
	}

	return parts[0], parts[1], nil
}

// format the strings into an id `a:b`
func BuildTwoPartID(a, b *string) string {
	return fmt.Sprintf("%s:%s", *a, *b)
}
