//go:build acceptance
// +build acceptance

package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_GitLab_ProviderMux(t *testing.T) {
	providerServer, err := NewMuxedProviderServer(context.Background(), "acctest")
	if err != nil {
		t.Fatalf("failed to create mux provider server for testing: %v", err)
	}
	//lintignore:AT001
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"gitlab": func() (tfprotov6.ProviderServer, error) { return providerServer(), nil },
		},
		Steps: []resource.TestStep{
			{
				Config: `
					// The gitlab_metadata data source is based on the terraform-plugin-framework
					data "gitlab_metadata" "test" {}
					
					// The gitlab_current_user data source is based on the terraform-plugin-sdk
					data "gitlab_current_user" "test" {}
                `,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify framework based data source attribute
					resource.TestCheckResourceAttr("data.gitlab_metadata.test", "id", "1"),
					// Verify sdk based data source attribute
					resource.TestCheckResourceAttr("data.gitlab_current_user.test", "id", "1"),
				),
			},
		},
	})
}
