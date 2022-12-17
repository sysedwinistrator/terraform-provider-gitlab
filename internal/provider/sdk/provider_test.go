//go:build acceptance
// +build acceptance

package sdk

import (
	"context"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/client"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/xanzy/go-gitlab"
)

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var providerFactoriesV6 = map[string]func() (tfprotov6.ProviderServer, error){
	"gitlab": func() (tfprotov6.ProviderServer, error) {
		provider, err := NewV6(context.Background(), "test")
		if err != nil {
			return nil, err
		}
		return provider, nil
	},
}

var testGitlabConfig = client.Config{
	Token:         os.Getenv("GITLAB_TOKEN"),
	BaseURL:       os.Getenv("GITLAB_BASE_URL"),
	CACertFile:    "",
	Insecure:      false,
	ClientCert:    "",
	ClientKey:     "",
	EarlyAuthFail: true,
}

var testGitlabClient *gitlab.Client

func init() {
	client, err := testGitlabConfig.NewGitLabClient(context.Background())
	if err != nil {
		panic("failed to create test client: " + err.Error()) // lintignore: R009 // TODO: Resolve this tfproviderlint issue
	}
	testGitlabClient = client
}

func TestProvider(t *testing.T) {
	t.Parallel()

	if err := New("dev")().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	t.Parallel()
	var _ *schema.Provider = New("dev")()
}
