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

func TestAccGitlabIntegrationSlack_basic(t *testing.T) {
	var slackService gitlab.SlackService
	rInt := acctest.RandInt()
	slackResourceName := "gitlab_integration_slack.slack"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		CheckDestroy:             testAccCheckGitlabServiceSlackDestroy,
		Steps: []resource.TestStep{
			// Create a project and a slack integration with minimal settings
			{
				Config: testAccGitlabIntegrationSlackMinimalConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabIntegrationExists(slackResourceName, &slackService),
					resource.TestCheckResourceAttr(slackResourceName, "webhook", "https://test.com"),
				),
			},
			{
				ResourceName:      slackResourceName,
				ImportStateIdFunc: getSlackProjectID(slackResourceName),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"notify_only_broken_pipelines",
					"notify_only_default_branch",
					"webhook",
				},
			},
			// Update slack integration with more settings
			{
				Config: testAccGitlabIntegrationSlackConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabIntegrationExists(slackResourceName, &slackService),
					resource.TestCheckResourceAttr(slackResourceName, "webhook", "https://test.com"),
					resource.TestCheckResourceAttr(slackResourceName, "push_events", "true"),
					resource.TestCheckResourceAttr(slackResourceName, "push_channel", "test"),
					// TODO: Currently, GitLab doesn't correctly implement the API, so this is
					//       impossible to implement here at the moment.
					//       see https://gitlab.com/gitlab-org/gitlab/-/issues/28903
					// resource.TestCheckResourceAttr(slackResourceName, "deployment_events", "true"),
					// resource.TestCheckResourceAttr(slackResourceName, "deployment_channel", "test"),
					resource.TestCheckResourceAttr(slackResourceName, "notify_only_broken_pipelines", "true"),
				),
			},
			{
				ResourceName:      slackResourceName,
				ImportStateIdFunc: getSlackProjectID(slackResourceName),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"notify_only_broken_pipelines",
					"notify_only_default_branch",
					"webhook",
				},
			},
			// Update the slack integration
			{
				Config: testAccGitlabIntegrationSlackUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabIntegrationExists(slackResourceName, &slackService),
					resource.TestCheckResourceAttr(slackResourceName, "webhook", "https://testwebhook.com"),
					resource.TestCheckResourceAttr(slackResourceName, "push_events", "false"),
					resource.TestCheckResourceAttr(slackResourceName, "push_channel", "test push_channel"),
					resource.TestCheckResourceAttr(slackResourceName, "notify_only_broken_pipelines", "false"),
				),
			},
			{
				ResourceName:      slackResourceName,
				ImportStateIdFunc: getSlackProjectID(slackResourceName),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"notify_only_broken_pipelines",
					"notify_only_default_branch",
					"webhook",
				},
			},
			// Update the slack integration to get back to previous settings
			{
				Config: testAccGitlabIntegrationSlackConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabIntegrationExists(slackResourceName, &slackService),
					resource.TestCheckResourceAttr(slackResourceName, "webhook", "https://test.com"),
					resource.TestCheckResourceAttr(slackResourceName, "push_events", "true"),
					resource.TestCheckResourceAttr(slackResourceName, "push_channel", "test"),
					resource.TestCheckResourceAttr(slackResourceName, "notify_only_broken_pipelines", "true"),
				),
			},
			{
				ResourceName:      slackResourceName,
				ImportStateIdFunc: getSlackProjectID(slackResourceName),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"notify_only_broken_pipelines",
					"notify_only_default_branch",
					"webhook",
				},
			},
			// Update the slack integration to get back to minimal settings
			{
				Config: testAccGitlabIntegrationSlackMinimalConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabIntegrationExists(slackResourceName, &slackService),
					resource.TestCheckResourceAttr(slackResourceName, "webhook", "https://test.com"),
					resource.TestCheckResourceAttr(slackResourceName, "push_channel", ""),
				),
			},
			// Verify Import
			{
				ResourceName:      slackResourceName,
				ImportStateIdFunc: getSlackProjectID(slackResourceName),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"notify_only_broken_pipelines",
					"notify_only_default_branch",
					"webhook",
				},
			},
		},
	})
}

func TestAccGitlabIntegrationSlack_backwardsCompatibility(t *testing.T) {
	var slackService gitlab.SlackService
	rInt := acctest.RandInt()
	slackResourceName := "gitlab_service_slack.slack"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		CheckDestroy:             testAccCheckGitlabServiceSlackDestroy,
		Steps: []resource.TestStep{
			// Create a project and a slack integration with minimal settings
			{
				Config: fmt.Sprintf(`
				resource "gitlab_project" "foo" {
				  name        = "foo-%d"
				  description = "Terraform acceptance tests"
				  visibility_level = "public"
				}
				
				resource "gitlab_service_slack" "slack" {
				  project                      = "${gitlab_project.foo.id}"
				  webhook                      = "https://test.com"
				}
				`, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabIntegrationExists(slackResourceName, &slackService),
					resource.TestCheckResourceAttr(slackResourceName, "webhook", "https://test.com"),
				),
			},
			{
				ResourceName:      slackResourceName,
				ImportStateIdFunc: getSlackProjectID(slackResourceName),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"notify_only_broken_pipelines",
					"notify_only_default_branch",
					"webhook",
				},
			},
		},
	})
}

func testAccCheckGitlabIntegrationExists(n string, service *gitlab.SlackService) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		project := rs.Primary.Attributes["project"]
		if project == "" {
			return fmt.Errorf("No project ID is set")
		}
		slackService, _, err := testutil.TestGitlabClient.Services.GetSlackService(project)
		if err != nil {
			return fmt.Errorf("Slack integration does not exist in project %s: %v", project, err)
		}
		*service = *slackService

		return nil
	}
}

func testAccCheckGitlabServiceSlackDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "gitlab_integration_slack" {
			continue
		}

		project := rs.Primary.ID

		_, _, err := testutil.TestGitlabClient.Services.GetSlackService(project)
		if err == nil {
			return fmt.Errorf("Slack Integration in project %s still exists", project)
		}
		if !api.Is404(err) {
			return err
		}
		return nil
	}
	return nil
}

func getSlackProjectID(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", n)
		}

		project := rs.Primary.Attributes["project"]
		if project == "" {
			return "", fmt.Errorf("No project ID is set")
		}

		return project, nil
	}
}

func testAccGitlabIntegrationSlackMinimalConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name        = "foo-%d"
  description = "Terraform acceptance tests"
  visibility_level = "public"
}

resource "gitlab_integration_slack" "slack" {
  project                      = "${gitlab_project.foo.id}"
  webhook                      = "https://test.com"
}
`, rInt)
}

func testAccGitlabIntegrationSlackConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name        = "foo-%d"
  description = "Terraform acceptance tests"
  visibility_level = "public"
}

resource "gitlab_integration_slack" "slack" {
  project                      = "${gitlab_project.foo.id}"
  webhook                      = "https://test.com"
  username                     = "test"
  push_events                  = true
  push_channel                 = "test"
  issues_events                = true
  issue_channel                = "test"
  confidential_issues_events   = true
  confidential_issue_channel   = "test"
  confidential_note_events     = true
// TODO: Currently, GitLab doesn't correctly implement the API, so this is
//       impossible to implement here at the moment.
//       see https://gitlab.com/gitlab-org/gitlab/-/issues/28903
//   deployment_channel           = "test"
//   deployment_events            = true
  merge_requests_events        = true
  merge_request_channel        = "test"
  tag_push_events              = true
  tag_push_channel             = "test"
  note_events                  = true
  note_channel                 = "test"
  pipeline_events              = true
  pipeline_channel             = "test"
  wiki_page_events             = true
  wiki_page_channel            = "test"
  notify_only_broken_pipelines = true
  branches_to_be_notified      = "all"
}
`, rInt)
}

func testAccGitlabIntegrationSlackUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name        = "foo-%d"
  description = "Terraform acceptance tests"
  visibility_level = "public"
}

resource "gitlab_integration_slack" "slack" {
  project                      = "${gitlab_project.foo.id}"
  webhook                      = "https://testwebhook.com"
  username                     = "test username"
  push_events                  = false
  push_channel                 = "test push_channel"
  issues_events                = false
  issue_channel                = "test issue_channel"
  confidential_issues_events   = false
  confidential_issue_channel   = "test confidential_issue_channel"
  confidential_note_events     = false
// TODO: Currently, GitLab doesn't correctly implement the API, so this is
//       impossible to implement here at the moment.
//       see https://gitlab.com/gitlab-org/gitlab/-/issues/28903
//   deployment_channel           = "test deployment_channel"
//   deployment_events            = false
  merge_requests_events        = false
  merge_request_channel        = "test merge_request_channel"
  tag_push_events              = false
  tag_push_channel             = "test tag_push_channel"
  note_events                  = false
  note_channel                 = "test note_channel"
  pipeline_events              = false
  pipeline_channel             = "test pipeline_channel"
  wiki_page_events             = false
  wiki_page_channel            = "test wiki_page_channel"
  notify_only_broken_pipelines = false
  branches_to_be_notified      = "all"
}
`, rInt)
}
