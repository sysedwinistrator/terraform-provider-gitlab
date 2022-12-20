//go:build acceptance
// +build acceptance

package sdk

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/testutil"
)

func TestAccDataGitlabProjectBranches_search(t *testing.T) {
	testProject := testutil.CreateProject(t)
	testBranches := testutil.CreateBranches(t, testProject, 25)
	expectedBranches := len(testBranches) + 1 //main branch already exists

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "gitlab_project_branches" "this" {
						project = "%d"
					}
				`, testProject.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gitlab_project_branches.this", "branches.#", fmt.Sprintf("%d", expectedBranches)),
					resource.TestCheckResourceAttrSet("data.gitlab_project_branches.this", "branches.0.name"),
					resource.TestCheckResourceAttrSet("data.gitlab_project_branches.this", "branches.0.merged"),
					resource.TestCheckResourceAttrSet("data.gitlab_project_branches.this", "branches.0.protected"),
					resource.TestCheckResourceAttrSet("data.gitlab_project_branches.this", "branches.0.default"),
					resource.TestCheckResourceAttrSet("data.gitlab_project_branches.this", "branches.0.developers_can_push"),
					resource.TestCheckResourceAttrSet("data.gitlab_project_branches.this", "branches.0.developers_can_merge"),
					resource.TestCheckResourceAttrSet("data.gitlab_project_branches.this", "branches.0.can_push"),
					resource.TestCheckResourceAttrSet("data.gitlab_project_branches.this", "branches.0.web_url"),
					resource.TestCheckResourceAttr("data.gitlab_project_branches.this", "branches.0.commit.#", "1"),
				),
			},
		},
	})
}

// This tests is testing that the update from https://github.com/mitchellh/hashstructure -> V2 maintains the ID properly
// and doesn't result in a breaking change.
func TestAccDataGitlabProjectBranches_UpdateHashStruct(t *testing.T) {
	testProject := testutil.CreateProject(t)

	// We want to use the same config on old and new version
	commonConfig := fmt.Sprintf(`
	data "gitlab_project_branches" "this" {
		project = "%d"
	}
	`, testProject.ID)

	oldID := ""

	resource.ParallelTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				//The version before this change was made.
				ExternalProviders: map[string]resource.ExternalProvider{
					"gitlab": {
						VersionConstraint: "3.20.0",
						Source:            "gitlabhq/gitlab",
					},
				},
				Config: commonConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.gitlab_project_branches.this", "id"),

					// Store the generated ID for the next step in the text.
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["data.gitlab_project_branches.this"]
						if !ok {
							return fmt.Errorf("data.gitlab_project_branches.this not found")
						}

						// Set the ID so we can check it later
						oldID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				ProtoV6ProviderFactories: providerFactoriesV6,
				Config:                   commonConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.gitlab_project_branches.this", "id"),

					// Verify that the old and new hashes match.
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["data.gitlab_project_branches.this"]
						if !ok {
							return fmt.Errorf("data.gitlab_project_branches.this not found")
						}

						//get the new ID to check it
						newID := rs.Primary.ID
						if newID != oldID {
							return fmt.Errorf("old and new IDs do not match! There is an error in the hash generation, likely in github.com/mitchellh/hashstructure/v2")
						}

						// hashes match.
						return nil
					},
				),
			},
		},
	})
}
