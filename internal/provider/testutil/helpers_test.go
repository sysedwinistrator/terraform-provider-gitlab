//go:build acceptance
// +build acceptance

package testutil

import (
	"github.com/xanzy/go-gitlab"
	"testing"
)

func TestIsRunningInEE(t *testing.T) {
	cases := []struct {
		name           string
		metadata       *gitlab.Metadata
		expectedResult bool
	}{
		{
			name: "EnterpriseIsTrue",
			metadata: &gitlab.Metadata{
				Enterprise: true,
			},
			expectedResult: true,
		},
		{
			name: "EnterpriseIsFalse",
			metadata: &gitlab.Metadata{
				Enterprise: false,
			},
			expectedResult: false,
		},
		{
			name: "EnterpriseIsFalseAndVersionIsEE",
			metadata: &gitlab.Metadata{
				Enterprise: false,
				Version:    "15.5.0-ee",
			},
			expectedResult: true,
		},
	}

	for _, tc := range cases {
		result := isEnterpriseInstance(tc.metadata)
		if result != tc.expectedResult {
			t.Fatalf("\"IsRunningInEE()\" FAILED, expected -> %v, got -> %v", tc.expectedResult, result)
		}
	}
}
