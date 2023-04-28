//go:build acceptance
// +build acceptance

package sdk

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/xanzy/go-gitlab"
	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/api"
	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/utils"

	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/testutil"
)

func TestAccGitlabGroupLdapLink_basicCN(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "gitlab_group_ldap_link.foo"

	// PreCheck runs after Config so load test data here
	var ldapLink gitlab.LDAPGroupLink
	testLdapLink := gitlab.LDAPGroupLink{
		CN:       "default",
		Provider: "default",
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		CheckDestroy:             testAccCheckGitlabGroupLdapLinkDestroy,
		Steps: []resource.TestStep{

			// Create a group LDAP link as a developer (uses testAccGitlabGroupLdapLinkCreateConfig for Config)
			{
				SkipFunc: testutil.IsRunningInCE,
				Config:   testAccGitlabGroupLdapLinkCreateConfig(rInt, &testLdapLink),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupLdapLinkExists(resourceName, &ldapLink),
					testAccCheckGitlabGroupLdapLinkAttributes(&ldapLink, &testAccGitlabGroupLdapLinkExpectedAttributes{
						accessLevel: "developer",
					})),
			},

			// Import the group LDAP link (re-uses testAccGitlabGroupLdapLinkCreateConfig for Config)
			{
				SkipFunc:          testutil.IsRunningInCE,
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force",
				},
			},

			// Update the group LDAP link to change the access level (uses testAccGitlabGroupLdapLinkUpdateConfig for Config)
			{
				SkipFunc: testutil.IsRunningInCE,
				Config:   testAccGitlabGroupLdapLinkUpdateConfig(rInt, &testLdapLink),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabGroupLdapLinkExists(resourceName, &ldapLink),
					testAccCheckGitlabGroupLdapLinkAttributes(&ldapLink, &testAccGitlabGroupLdapLinkExpectedAttributes{
						accessLevel: "maintainer",
					})),
			},
		},
	})
}

func TestAccGitlabGroupLdapLink_basicFilter(t *testing.T) {
	resourceName := "gitlab_group_ldap_link.foo"

	group := testutil.CreateGroups(t, 1)[0]

	// PreCheck runs after Config so load test data here
	var ldapLink gitlab.LDAPGroupLink

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		CheckDestroy:             testAccCheckGitlabGroupLdapLinkDestroy,
		Steps: []resource.TestStep{

			// Create a group LDAP link using a valid filter
			{
				SkipFunc: testutil.IsRunningInCE,
				Config: fmt.Sprintf(`resource "gitlab_group_ldap_link" "foo" {
					group_id 		= "%d"
					filter          = "(&(objectClass=person)(objectClass=user))"
					group_access 	= "developer"
					ldap_provider   = "default"
				
				}`, group.ID),
				Check: testAccCheckGitlabGroupLdapLinkExists(resourceName, &ldapLink),
			},
			{
				SkipFunc:          testutil.IsRunningInCE,
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force",
				},
			},
		},
	})
}

func TestAccGitlabGroupLdapLink_conflictingArguments(t *testing.T) {
	group := testutil.CreateGroups(t, 1)[0]

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		CheckDestroy:             testAccCheckGitlabGroupLdapLinkDestroy,
		Steps: []resource.TestStep{

			// Create a group LDAP link using conflicting arguments
			// ensure both conflict errors are printed appropriately.
			{
				SkipFunc: testutil.IsRunningInCE,
				Config: fmt.Sprintf(`resource "gitlab_group_ldap_link" "foo" {
					group_id 		= "%d"
					cn              = "default"
					filter          = "(&(objectClass=person)(objectClass=user))"
					group_access 	= "developer"
					ldap_provider   = "default"
				}`, group.ID),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta(`"cn": conflicts with filter`)),
			},
			{
				SkipFunc: testutil.IsRunningInCE,
				Config: fmt.Sprintf(`resource "gitlab_group_ldap_link" "foo" {
					group_id 		= "%d"
					cn              = "default"
					filter          = "(&(objectClass=person)(objectClass=user))"
					group_access 	= "developer"
					ldap_provider   = "default"
				}`, group.ID),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta(`"filter": conflicts with cn`)),
			},
		},
	})
}

func TestAccGitlabGroupLdapLink_recreatedWhenRemoved(t *testing.T) {
	testutil.SkipIfCE(t)

	testGroup := testutil.CreateGroups(t, 1)[0]
	ldapName := acctest.RandomWithPrefix("ldap")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		CheckDestroy:             testAccCheckGitlabGroupLdapLinkDestroy,
		Steps: []resource.TestStep{
			// Create an LDAP group link
			{
				Config: fmt.Sprintf(`
          resource "gitlab_group_ldap_link" "test" {
            group_id      = "%[1]d"
            cn            = "%[2]s"
            group_access  = "developer"
            ldap_provider = "%[2]s"
          }
          `, testGroup.ID, ldapName),
			},
			// Verify Import
			{
				ResourceName:      "gitlab_group_ldap_link.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force",
				},
			},
			// Remove the LDAP link directly and re-apply the config to re-create the LDAP link
			{
				PreConfig: func() {
					if _, err := testutil.TestGitlabClient.Groups.DeleteGroupLDAPLink(testGroup.ID, ldapName); err != nil {
						t.Fatalf("Failed to delete LDAP link %q in group %d", ldapName, testGroup.ID)
					}
				},
				Config: fmt.Sprintf(`
          resource "gitlab_group_ldap_link" "test" {
            group_id      = "%[1]d"
            cn            = "%[2]s"
            group_access  = "developer"
            ldap_provider = "%[2]s"
          }
          `, testGroup.ID, ldapName),
			},
			// Verify Import
			{
				ResourceName:      "gitlab_group_ldap_link.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force",
				},
			},
		},
	})
}

func TestAccGitlabGroupLdapLink_StateUpgradeV0(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name            string
		givenV0State    map[string]interface{}
		expectedV1State map[string]interface{}
	}{
		{
			name: "Project With ID and CN",
			givenV0State: map[string]interface{}{
				"group_id":      "99",
				"cn":            "mainScreenTurnOn",
				"ldap_provider": "allYourBase",
				"filter":        "",
				"id":            "allYourBase:mainScreenTurnOn",
			},
			expectedV1State: map[string]interface{}{
				"group_id":      "99",
				"cn":            "mainScreenTurnOn",
				"ldap_provider": "allYourBase",
				"filter":        "",
				"id":            "99:allYourBase:mainScreenTurnOn:",
			},
		},
		{
			name: "Project With ID and Filter",
			givenV0State: map[string]interface{}{
				"group_id":      "99",
				"cn":            "",
				"filter":        "thisIsAFilter",
				"ldap_provider": "allYourBase",
				"id":            "allYourBase:mainScreenTurnOn",
			},
			expectedV1State: map[string]interface{}{
				"group_id":      "99",
				"cn":            "",
				"filter":        "thisIsAFilter",
				"ldap_provider": "allYourBase",
				"id":            "99:allYourBase::thisIsAFilter",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualV1State, err := resourceGitlabGroupLDAPLinkStateUpgradeV0(context.Background(), tc.givenV0State, nil)
			if err != nil {
				t.Fatalf("Error migrating state: %s", err)
			}

			if !reflect.DeepEqual(tc.expectedV1State, actualV1State) {
				t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", tc.expectedV1State, actualV1State)
			}
		})

	}
}

func testAccCheckGitlabGroupLdapLinkExists(resourceName string, ldapLink *gitlab.LDAPGroupLink) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Clear the "found" LDAP link before checking for existence
		*ldapLink = gitlab.LDAPGroupLink{}

		resourceState, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		err := testAccGetGitlabGroupLdapLink(ldapLink, resourceState)
		if err != nil {
			return err
		}

		return nil
	}
}

type testAccGitlabGroupLdapLinkExpectedAttributes struct {
	accessLevel string
}

func testAccCheckGitlabGroupLdapLinkAttributes(ldapLink *gitlab.LDAPGroupLink, want *testAccGitlabGroupLdapLinkExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		accessLevelId, ok := api.AccessLevelValueToName[ldapLink.GroupAccess]
		if !ok {
			return fmt.Errorf("Invalid access level '%s'", accessLevelId)
		}
		if accessLevelId != want.accessLevel {
			return fmt.Errorf("Has access level %s; want %s", accessLevelId, want.accessLevel)
		}
		return nil
	}
}

func testAccCheckGitlabGroupLdapLinkDestroy(s *terraform.State) error {
	// Can't check for links if the group is destroyed so make sure all groups are destroyed instead
	for _, resourceState := range s.RootModule().Resources {
		if resourceState.Type != "gitlab_group" {
			continue
		}

		group, _, err := testutil.TestGitlabClient.Groups.GetGroup(resourceState.Primary.ID, nil)
		if err == nil {
			if group != nil && fmt.Sprintf("%d", group.ID) == resourceState.Primary.ID {
				if group.MarkedForDeletionOn == nil {
					return fmt.Errorf("Group still exists")
				}
			}
		}
		if !api.Is404(err) {
			return err
		}
		return nil
	}
	return nil
}

func testAccGetGitlabGroupLdapLink(ldapLink *gitlab.LDAPGroupLink, resourceState *terraform.ResourceState) error {
	groupId := resourceState.Primary.Attributes["group_id"]
	if groupId == "" {
		return fmt.Errorf("No group ID is set")
	}

	// Construct our desired LDAP Link from the config values
	desiredLdapLink := gitlab.LDAPGroupLink{
		CN:          resourceState.Primary.Attributes["cn"],
		GroupAccess: api.AccessLevelNameToValue[resourceState.Primary.Attributes["group_access"]],
		Provider:    resourceState.Primary.Attributes["ldap_provider"],
	}

	desiredLdapLinkId := utils.BuildTwoPartID(&desiredLdapLink.Provider, &desiredLdapLink.CN)

	// Try to fetch all group links from GitLab
	currentLdapLinks, _, err := testutil.TestGitlabClient.Groups.ListGroupLDAPLinks(groupId, nil)
	if err != nil {
		// The read/GET API wasn't implemented in GitLab until version 12.8 (March 2020, well after the add and delete APIs).
		// If we 404, assume GitLab is at an older version and take things on faith.
		switch err.(type) { // nolint // TODO: Resolve this golangci-lint issue: S1034: assigning the result of this type assertion to a variable (switch err := err.(type)) could eliminate type assertions in switch cases (gosimple)
		case *gitlab.ErrorResponse:
			if err.(*gitlab.ErrorResponse).Response.StatusCode == 404 { // nolint // TODO: Resolve this golangci-lint issue: S1034(related information): could eliminate this type assertion (gosimple)
				// Do nothing
			} else {
				return err
			}
		default:
			return err
		}
	}

	// If we got here and don't have links, assume GitLab is below version 12.8 and skip the check
	if currentLdapLinks != nil {
		found := false

		// Check if the LDAP link exists in the returned list of links
		for _, currentLdapLink := range currentLdapLinks {
			if utils.BuildTwoPartID(&currentLdapLink.Provider, &currentLdapLink.CN) == desiredLdapLinkId {
				found = true
				*ldapLink = *currentLdapLink
				break
			}
		}

		if !found {
			return errors.New(fmt.Sprintf("LdapLink %s does not exist.", desiredLdapLinkId)) // nolint // TODO: Resolve this golangci-lint issue: S1028: should use fmt.Errorf(...) instead of errors.New(fmt.Sprintf(...)) (gosimple)
		}
	} else {
		*ldapLink = desiredLdapLink
	}

	return nil
}

func testAccGitlabGroupLdapLinkCreateConfig(rInt int, testLdapLink *gitlab.LDAPGroupLink) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
    name = "foo%d"
	path = "foo%d"
	description = "Terraform acceptance test - Group LDAP Links 1"
}

resource "gitlab_group_ldap_link" "foo" {
    group_id 		= "${gitlab_group.foo.id}"
    cn				= "%s"
	group_access 	= "developer"
	ldap_provider   = "%s"

}`, rInt, rInt, testLdapLink.CN, testLdapLink.Provider)
}

func testAccGitlabGroupLdapLinkUpdateConfig(rInt int, testLdapLink *gitlab.LDAPGroupLink) string {
	return fmt.Sprintf(`
resource "gitlab_group" "foo" {
    name = "foo%d"
	path = "foo%d"
	description = "Terraform acceptance test - Group LDAP Links 2"
}

resource "gitlab_group_ldap_link" "foo" {
    group_id 		= "${gitlab_group.foo.id}"
    cn				= "%s"
	group_access 	= "maintainer"
	ldap_provider   = "%s"
}`, rInt, rInt, testLdapLink.CN, testLdapLink.Provider)
}
