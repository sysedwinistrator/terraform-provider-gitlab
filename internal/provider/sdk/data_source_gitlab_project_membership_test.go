//go:build acceptance
// +build acceptance

package sdk

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/testutil"
)

func TestAccDataSourceGitlabProjectMembership_basic(t *testing.T) {

	project := testutil.CreateProject(t)
	users := testutil.CreateUsers(t, 1)
	testutil.AddProjectMembers(t, project.ID, users)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGitlabProjectMembership(project.ID),
				Check: resource.ComposeTestCheckFunc(
					// Members is 2 because the user owning the token is always added to the project
					resource.TestCheckResourceAttr("data.gitlab_project_membership.foo", "members.#", "2"),
					resource.TestCheckResourceAttr("data.gitlab_project_membership.foo", "members.1.username", users[0].Username),
					resource.TestCheckResourceAttr("data.gitlab_project_membership.foo", "members.1.access_level", "developer"),
				),
			},
		},
	})
}

func TestAccDataSourceGitlabProjectMembership_pagination(t *testing.T) {
	userCount := 21

	project := testutil.CreateProject(t)
	users := testutil.CreateUsers(t, userCount)
	testutil.AddProjectMembers(t, project.ID, users)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGitlabProjectMembership(project.ID),
				// one more for the user owning the token, which is always added to the project.
				Check: resource.TestCheckResourceAttr("data.gitlab_project_membership.foo", "members.#", fmt.Sprintf("%d", userCount+1)),
			},
		},
	})
}

func testAccDataSourceGitlabProjectMembership(projectID int) string {
	return fmt.Sprintf(`
data "gitlab_project_membership" "foo" {
  project_id = "%d"
}`, projectID)
}
