package utils

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"regexp"
)

// HttpUrlValidator validates that URL starts with http or https schema
var HttpUrlValidator = stringvalidator.RegexMatches(regexp.MustCompile(`^https?://`), "value should be an URL with http or https schema")
