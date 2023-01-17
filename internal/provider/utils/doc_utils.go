package utils

import (
	"fmt"
	"strings"
)

func RenderValueListForDocs(values []string) string {
	inlineCodeValues := make([]string, 0, len(values))
	for _, v := range values {
		inlineCodeValues = append(inlineCodeValues, fmt.Sprintf("`%s`", v))
	}
	return strings.Join(inlineCodeValues, ", ")
}
