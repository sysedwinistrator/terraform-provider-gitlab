//go:build acceptance
// +build acceptance

package sdk

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/testutil"
)

func TestAccDataSourceGitlabGroupMembership_basic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		Steps: []resource.TestStep{
			// Create the group and one member
			{
				Config: testAccDataSourceGitlabGroupMembershipConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("gitlab_group.foo", "name", fmt.Sprintf("foo%d", rInt)),
					resource.TestCheckResourceAttr("gitlab_user.test", "name", fmt.Sprintf("foo%d", rInt)),
					resource.TestCheckResourceAttr("gitlab_group_membership.foo", "access_level", "developer"),
				),
			},
			{
				Config: testAccDataSourceGitlabGroupMembershipConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					// Members is 2 because the user owning the token is always added to the group
					resource.TestCheckResourceAttr("data.gitlab_group_membership.foo", "members.#", "2"),
					resource.TestCheckResourceAttr("data.gitlab_group_membership.foo", "members.1.username", fmt.Sprintf("listest%d", rInt)),
				),
			},

			// Get group using its ID, but return maintainers only
			{
				Config: testAccDataSourceGitlabGroupMembershipConfigFilterAccessLevel(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.gitlab_group_membership.foomaintainers", "members.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceGitlabGroupMembership_inherited(t *testing.T) {
	// create the parent group
	parentGroup := testutil.CreateGroups(t, 1)[0]
	// create the nested group
	nestedGroup := testutil.CreateSubGroups(t, parentGroup, 1)[0]
	// create user
	user := testutil.CreateUsers(t, 1)
	// add user to the parent_group (will be added as Developer)
	testutil.AddGroupMembers(t, parentGroup.ID, user)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "gitlab_group_membership" "this" {
					  group_id     = "%d"
					  access_level = "developer"
					  inherited = true
					}`, nestedGroup.ID),
				Check: resource.TestCheckResourceAttr("data.gitlab_group_membership.this", "members.0.username", user[0].Username),
			},
		},
	})
}

func TestAccDataSourceGitlabGroupMembership_pagination(t *testing.T) {
	userCount := 21

	group := testutil.CreateGroups(t, 1)[0]
	users := testutil.CreateUsers(t, userCount)
	testutil.AddGroupMembers(t, group.ID, users)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGitlabGroupMembershipPagination(group.ID),
				Check:  resource.TestCheckResourceAttr("data.gitlab_group_membership.this", "members.#", fmt.Sprintf("%d", userCount)),
			},
		},
	})
}

func testAccDataSourceGitlabGroupMembershipConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
  name = "foo%d"
  path = "foo%d"
}

resource "gitlab_user" "test" {
  name     = "foo%d"
  username = "listest%d"
  password = "%s"
  email    = "listest%d@ssss.com"
}

resource "gitlab_group_membership" "foo" {
  group_id     = "${gitlab_group.foo.id}"
  user_id      = "${gitlab_user.test.id}"
  access_level = "developer"
}`, rInt, rInt, rInt, rInt, acctest.RandString(16), rInt)
}

func testAccDataSourceGitlabGroupMembershipConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
  name = "foo%d"
  path = "foo%d"
}

data "gitlab_group_membership" "foo" {
  group_id = "${gitlab_group.foo.id}"
}`, rInt, rInt)
}

func testAccDataSourceGitlabGroupMembershipConfigFilterAccessLevel(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
  name = "foo%d"
  path = "foo%d"
}

data "gitlab_group_membership" "foomaintainers" {
  group_id     = "${gitlab_group.foo.id}"
  access_level = "maintainer"
}`, rInt, rInt)
}

func testAccDataSourceGitlabGroupMembershipPagination(groupId int) string {
	return fmt.Sprintf(`
data "gitlab_group_membership" "this" {
  group_id     = "%d"
  access_level = "developer"
}`, groupId)
}
