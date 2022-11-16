//go:build acceptance
// +build acceptance

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataGitlabProjectBranches_search(t *testing.T) {
	testProject := testAccCreateProject(t)
	testBranches := testAccCreateBranches(t, testProject, 25)
	expectedBranches := len(testBranches) + 1 //main branch already exists

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: providerFactories,
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
