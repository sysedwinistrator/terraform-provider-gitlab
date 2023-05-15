//go:build acceptance
// +build acceptance

package sdk

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/xanzy/go-gitlab"
	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/utils"

	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/testutil"
)

func TestAccGitlabGroupAccessToken_basic(t *testing.T) {
	var gat testAccGitlabGroupAccessTokenWrapper
	var groupVariable gitlab.GroupVariable

	testGroup := testutil.CreateGroups(t, 1)[0]

	expiresAt := time.Now().AddDate(0, 1, 0)
	updatedExpiresAt := expiresAt.AddDate(0, 1, 0)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		CheckDestroy:             testAccCheckGitlabGroupAccessTokenDestroy,
		Steps: []resource.TestStep{
			// Create a Group and a Group Access Token
			{
				Config: testAccGitlabGroupAccessTokenConfig(testGroup.ID, expiresAt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupAccessTokenExists("gitlab_group_access_token.this", &gat),
					testAccCheckGitlabGroupAccessTokenAttributes(&gat, &testAccGitlabGroupAccessTokenExpectedAttributes{
						name:        "my group token",
						scopes:      map[string]bool{"read_repository": true, "api": true, "write_repository": true, "read_api": true},
						expiresAt:   expiresAt.Format(iso8601),
						accessLevel: gitlab.AccessLevelValue(gitlab.DeveloperPermissions),
					}),
				),
			},
			// Update the Group Access Token to change the parameters
			{
				Config: testAccGitlabGroupAccessTokenUpdateConfig(testGroup.ID, updatedExpiresAt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupAccessTokenExists("gitlab_group_access_token.this", &gat),
					testAccCheckGitlabGroupAccessTokenAttributes(&gat, &testAccGitlabGroupAccessTokenExpectedAttributes{
						name:        "my new group token",
						scopes:      map[string]bool{"read_repository": false, "api": true, "write_repository": false, "read_api": false},
						expiresAt:   updatedExpiresAt.Format(iso8601),
						accessLevel: gitlab.AccessLevelValue(gitlab.MaintainerPermissions),
					}),
				),
			},
			// Update the Group Access Token Access Level to Owner
			{
				Config: testAccGitlabGroupAccessTokenUpdateAccessLevel(testGroup.ID, updatedExpiresAt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupAccessTokenExists("gitlab_group_access_token.this", &gat),
					testAccCheckGitlabGroupAccessTokenAttributes(&gat, &testAccGitlabGroupAccessTokenExpectedAttributes{
						name:        "my new group token",
						scopes:      map[string]bool{"read_repository": false, "api": true, "write_repository": false, "read_api": false},
						expiresAt:   updatedExpiresAt.Format(iso8601),
						accessLevel: gitlab.AccessLevelValue(gitlab.OwnerPermissions),
					}),
				),
			},
			// Add a CICD variable with Group Access Token value
			{
				Config: testAccGitlabGroupAccessTokenUpdateConfigWithCICDvar(testGroup.ID, updatedExpiresAt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupAccessTokenExists("gitlab_group_access_token.this", &gat),
					testAccCheckGitlabGroupVariableExists("gitlab_group_variable.var", &groupVariable),
					testAccCheckGitlabGroupAccessTokenAttributes(&gat, &testAccGitlabGroupAccessTokenExpectedAttributes{
						name:        "my new group token",
						scopes:      map[string]bool{"read_repository": false, "api": true, "write_repository": false, "read_api": false},
						expiresAt:   updatedExpiresAt.Format(iso8601),
						accessLevel: gitlab.AccessLevelValue(gitlab.MaintainerPermissions),
					}),
				),
			},
			//Restore Group Access Token initial parameters
			{
				Config: testAccGitlabGroupAccessTokenConfig(testGroup.ID, expiresAt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupAccessTokenExists("gitlab_group_access_token.this", &gat),
					testAccCheckGitlabGroupAccessTokenAttributes(&gat, &testAccGitlabGroupAccessTokenExpectedAttributes{
						name:        "my group token",
						scopes:      map[string]bool{"read_repository": true, "api": true, "write_repository": true, "read_api": true},
						expiresAt:   expiresAt.Format(iso8601),
						accessLevel: gitlab.AccessLevelValue(gitlab.DeveloperPermissions),
					}),
				),
			},
			// Verify import
			{
				ResourceName:      "gitlab_group_access_token.this",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// the token is only known during creating. We explicitly mention this limitation in the docs.
					"token",
				},
			},
		},
	})
}

func testAccCheckGitlabGroupAccessTokenExists(n string, gat *testAccGitlabGroupAccessTokenWrapper) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		group, tokenString, err := utils.ParseTwoPartID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error parsing ID: %s", rs.Primary.ID)
		}
		groupAccessTokenID, err := strconv.Atoi(tokenString)
		if err != nil {
			return fmt.Errorf("%s cannot be converted to int", tokenString)
		}

		groupId := rs.Primary.Attributes["group"]
		if groupId == "" {
			return fmt.Errorf("No group ID is set")
		}
		if groupId != group {
			return fmt.Errorf("Group [%s] in group identifier [%s] it's different from group stored into the state [%s]", group, rs.Primary.ID, groupId)
		}

		tokens, _, err := testutil.TestGitlabClient.GroupAccessTokens.ListGroupAccessTokens(groupId, nil)
		if err != nil {
			return err
		}

		for _, token := range tokens {
			if token.ID == groupAccessTokenID {
				gat.groupAccessToken = token
				gat.group = groupId
				gat.token = rs.Primary.Attributes["token"]
				return nil
			}
		}
		return fmt.Errorf("Group Access Token does not exist")
	}
}

type testAccGitlabGroupAccessTokenExpectedAttributes struct {
	name        string
	scopes      map[string]bool
	expiresAt   string
	accessLevel gitlab.AccessLevelValue
}

type testAccGitlabGroupAccessTokenWrapper struct {
	groupAccessToken *gitlab.GroupAccessToken
	group            string
	token            string
}

func testAccCheckGitlabGroupAccessTokenAttributes(gatWrap *testAccGitlabGroupAccessTokenWrapper, want *testAccGitlabGroupAccessTokenExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		gat := gatWrap.groupAccessToken
		if gat.Name != want.name {
			return fmt.Errorf("got Name %q; want %q", gat.Name, want.name)
		}

		if gat.ExpiresAt.String() != want.expiresAt {
			return fmt.Errorf("got ExpiresAt %q; want %q", gat.ExpiresAt.String(), want.expiresAt)
		}

		if gat.AccessLevel != want.accessLevel {
			return fmt.Errorf("got AccessLevel %q; want %q", gat.AccessLevel, want.accessLevel)
		}

		for _, scope := range gat.Scopes {
			if !want.scopes[scope] {
				return fmt.Errorf("got a not wanted Scope %q, received %v", scope, gat.Scopes)
			}
			want.scopes[scope] = false
		}
		for k, v := range want.scopes {
			if v {
				return fmt.Errorf("not got a wanted Scope %q, received %v", k, gat.Scopes)
			}
		}

		git, err := gitlab.NewClient(gatWrap.token, gitlab.WithBaseURL(testutil.TestGitlabClient.BaseURL().String()))
		if err != nil {
			return fmt.Errorf("Cannot use the token to instantiate a new client %s", err)
		}
		_, _, err = git.Groups.GetGroup(gatWrap.group, nil)
		if err != nil {
			return fmt.Errorf("Cannot use the token to perform an API call %s", err)
		}

		return nil
	}
}

func testAccCheckGitlabGroupAccessTokenDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "gitlab_group" {
			continue
		}

		group, resp, err := testutil.TestGitlabClient.Groups.GetGroup(rs.Primary.ID, nil)
		if err == nil {
			if group != nil && fmt.Sprintf("%d", group.ID) == rs.Primary.ID {
				if group.MarkedForDeletionOn == nil {
					return fmt.Errorf("Group still exists")
				}
			}
		}
		if resp.StatusCode != 404 {
			return err
		}
		return nil
	}
	return nil
}

func testAccGitlabGroupAccessTokenConfig(groupId int, expiresAt time.Time) string {
	return fmt.Sprintf(`
resource "gitlab_group_access_token" "this" {
  name = "my group token"
  group = %d
  expires_at = "%s"
  access_level = "developer"
  scopes = ["read_repository" , "api", "write_repository", "read_api"]
}
	`, groupId, expiresAt.Format(iso8601))
}

func testAccGitlabGroupAccessTokenUpdateConfig(groupId int, expiresAt time.Time) string {
	return fmt.Sprintf(`
resource "gitlab_group_access_token" "this" {
  name = "my new group token"
  group = %d
  expires_at = "%s"
  access_level = "maintainer"
  scopes = ["api"]
}
	`, groupId, expiresAt.Format(iso8601))
}

func testAccGitlabGroupAccessTokenUpdateAccessLevel(groupId int, expiresAt time.Time) string {
	return fmt.Sprintf(`
resource "gitlab_group_access_token" "this" {
  name = "my new group token"
  group = %d
  expires_at = "%s"
  access_level = "owner"
  scopes = ["api"]
}
	`, groupId, expiresAt.Format(iso8601))
}

func testAccGitlabGroupAccessTokenUpdateConfigWithCICDvar(groupId int, expiresAt time.Time) string {
	return fmt.Sprintf(`
resource "gitlab_group_access_token" "this" {
  name = "my new group token"
  group = %[1]d
  expires_at = "%[2]s"
  access_level = "maintainer"
  scopes = ["api"]
}

resource "gitlab_group_variable" "var" {
  group   = %[1]d
  key     = "my_grp_access_token"
  value   = gitlab_group_access_token.this.token
 }

	`, groupId, expiresAt.Format(iso8601))
}
