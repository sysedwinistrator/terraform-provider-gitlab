//go:build acceptance
// +build acceptance

package sdk

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/xanzy/go-gitlab"
	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/api"

	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/testutil"
)

func TestAcc_GitlabIntegrationJira_basic(t *testing.T) {
	var jiraService gitlab.JiraService
	rInt := acctest.RandInt()
	jiraResourceName := "gitlab_integration_jira.jira"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		CheckDestroy:             testAccCheckGitlabIntegrationJiraDestroy,
		Steps: []resource.TestStep{
			// Create a project and a jira service
			{
				Config: testAccGitlabIntegrationJiraConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabIntegrationJiraExists(jiraResourceName, &jiraService),
					resource.TestCheckResourceAttr(jiraResourceName, "url", "https://test.com"),
					resource.TestCheckResourceAttr(jiraResourceName, "username", "user1"),
					resource.TestCheckResourceAttr(jiraResourceName, "password", "mypass"),
					resource.TestCheckResourceAttr(jiraResourceName, "commit_events", "true"),
					resource.TestCheckResourceAttr(jiraResourceName, "merge_requests_events", "false"),
					resource.TestCheckResourceAttr(jiraResourceName, "comment_on_event_enabled", "false"),
				),
			},
			// Verify Import
			{
				ResourceName:      jiraResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
			// Update the jira service
			{
				Config: testAccGitlabIntegrationJiraUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabIntegrationJiraExists(jiraResourceName, &jiraService),
					resource.TestCheckResourceAttr(jiraResourceName, "url", "https://testurl.com"),
					resource.TestCheckResourceAttr(jiraResourceName, "api_url", "https://testurl.com/rest"),
					resource.TestCheckResourceAttr(jiraResourceName, "username", "user2"),
					resource.TestCheckResourceAttr(jiraResourceName, "password", "mypass_update"),
					resource.TestCheckResourceAttr(jiraResourceName, "jira_issue_transition_id", "3"),
					resource.TestCheckResourceAttr(jiraResourceName, "commit_events", "false"),
					resource.TestCheckResourceAttr(jiraResourceName, "merge_requests_events", "true"),
					resource.TestCheckResourceAttr(jiraResourceName, "comment_on_event_enabled", "true"),
				),
			},
			// Verify Import
			{
				ResourceName:      jiraResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
			// Update the jira service to get back to previous settings
			{
				Config: testAccGitlabIntegrationJiraConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabIntegrationJiraExists(jiraResourceName, &jiraService),
					resource.TestCheckResourceAttr(jiraResourceName, "url", "https://test.com"),
					resource.TestCheckResourceAttr(jiraResourceName, "api_url", "https://testurl.com/rest"),
					resource.TestCheckResourceAttr(jiraResourceName, "username", "user1"),
					resource.TestCheckResourceAttr(jiraResourceName, "password", "mypass"),
					resource.TestCheckResourceAttr(jiraResourceName, "commit_events", "true"),
					resource.TestCheckResourceAttr(jiraResourceName, "merge_requests_events", "false"),
					resource.TestCheckResourceAttr(jiraResourceName, "comment_on_event_enabled", "false"),
				),
			},
			// Verify Import
			{
				ResourceName:      jiraResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAcc_GitlabIntegrationJira_backwardsCompatibility(t *testing.T) {
	var jiraService gitlab.JiraService
	rInt := acctest.RandInt()
	jiraResourceName := "gitlab_service_jira.jira"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		CheckDestroy:             testAccCheckGitlabIntegrationJiraDestroy,
		Steps: []resource.TestStep{
			// Create a project and a jira service
			{
				Config: fmt.Sprintf(`
				resource "gitlab_project" "foo" {
				  name        = "foo-%d"
				  description = "Terraform acceptance tests"
				  visibility_level = "public"
				}
				
				resource "gitlab_service_jira" "jira" {
				  project  = "${gitlab_project.foo.id}"
				  url      = "https://test.com"
				  username = "user1"
				  password = "mypass"
				  commit_events = true
				  merge_requests_events    = false
				  comment_on_event_enabled = false
				}
				`, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabIntegrationJiraExists(jiraResourceName, &jiraService),
					resource.TestCheckResourceAttr(jiraResourceName, "url", "https://test.com"),
					resource.TestCheckResourceAttr(jiraResourceName, "username", "user1"),
					resource.TestCheckResourceAttr(jiraResourceName, "password", "mypass"),
					resource.TestCheckResourceAttr(jiraResourceName, "commit_events", "true"),
					resource.TestCheckResourceAttr(jiraResourceName, "merge_requests_events", "false"),
					resource.TestCheckResourceAttr(jiraResourceName, "comment_on_event_enabled", "false"),
				),
			},
			// Verify Import
			{
				ResourceName:      jiraResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func testAccCheckGitlabIntegrationJiraExists(n string, service *gitlab.JiraService) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		project := rs.Primary.Attributes["project"]
		if project == "" {
			return fmt.Errorf("No project ID is set")
		}
		jiraService, _, err := testutil.TestGitlabClient.Services.GetJiraService(project)
		if err != nil {
			return fmt.Errorf("Jira integration does not exist in project %s: %v", project, err)
		}
		*service = *jiraService

		return nil
	}
}

func testAccCheckGitlabIntegrationJiraDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "gitlab_integration_jira" {
			continue
		}

		project := rs.Primary.ID

		_, _, err := testutil.TestGitlabClient.Services.GetJiraService(project)
		if err == nil {
			return fmt.Errorf("Jira Integration in project %s still exists", project)
		}
		if !api.Is404(err) {
			return err
		}
		return nil
	}
	return nil
}

func testAccGitlabIntegrationJiraConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name        = "foo-%d"
  description = "Terraform acceptance tests"
  visibility_level = "public"
}

resource "gitlab_integration_jira" "jira" {
  project  = "${gitlab_project.foo.id}"
  url      = "https://test.com"
  username = "user1"
  password = "mypass"
  commit_events = true
  merge_requests_events    = false
  comment_on_event_enabled = false
}
`, rInt)
}

func testAccGitlabIntegrationJiraUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name        = "foo-%d"
  description = "Terraform acceptance tests"
  visibility_level = "public"
}

resource "gitlab_integration_jira" "jira" {
  project  = "${gitlab_project.foo.id}"
  url      = "https://testurl.com"
  api_url  = "https://testurl.com/rest"
  username = "user2"
  password = "mypass_update"
  jira_issue_transition_id = "3"
  commit_events = false
  merge_requests_events    = true
  comment_on_event_enabled = true
}
`, rInt)
}
