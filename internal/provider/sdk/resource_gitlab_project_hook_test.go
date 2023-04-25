//go:build acceptance
// +build acceptance

package sdk

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/xanzy/go-gitlab"
	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/api"

	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/testutil"
)

func TestAccGitlabProjectHook_basic(t *testing.T) {
	var hook gitlab.ProjectHook
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		CheckDestroy:             testAccCheckGitlabProjectHookDestroy,
		Steps: []resource.TestStep{
			// Create a project and hook with default options
			{
				Config: testAccGitlabProjectHookConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectHookExists("gitlab_project_hook.foo", &hook),
					testAccCheckGitlabProjectHookAttributes(&hook, &testAccGitlabProjectHookExpectedAttributes{
						URL:                   fmt.Sprintf("https://example.com/hook-%d", rInt),
						PushEvents:            true,
						EnableSSLVerification: true,
					}),
				),
			},
			// Update the project hook to toggle all the values to their inverse
			{
				Config: testAccGitlabProjectHookUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectHookExists("gitlab_project_hook.foo", &hook),
					testAccCheckGitlabProjectHookAttributes(&hook, &testAccGitlabProjectHookExpectedAttributes{
						URL:                      fmt.Sprintf("https://example.com/hook-%d", rInt),
						PushEvents:               true,
						PushEventsBranchFilter:   "devel",
						IssuesEvents:             false,
						ConfidentialIssuesEvents: false,
						MergeRequestsEvents:      true,
						TagPushEvents:            true,
						NoteEvents:               true,
						ConfidentialNoteEvents:   true,
						JobEvents:                true,
						PipelineEvents:           true,
						WikiPageEvents:           true,
						DeploymentEvents:         true,
						ReleasesEvents:           true,
						EnableSSLVerification:    false,
					}),
				),
			},
			// Update the project hook to toggle the options back
			{
				Config: testAccGitlabProjectHookConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabProjectHookExists("gitlab_project_hook.foo", &hook),
					testAccCheckGitlabProjectHookAttributes(&hook, &testAccGitlabProjectHookExpectedAttributes{
						URL:                   fmt.Sprintf("https://example.com/hook-%d", rInt),
						PushEvents:            true,
						EnableSSLVerification: true,
					}),
				),
			},
			// Verify import
			{
				ResourceName:            "gitlab_project_hook.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func testAccCheckGitlabProjectHookExists(n string, hook *gitlab.ProjectHook) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		project, hookID, err := resourceGitlabProjectHookParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		gotHook, _, err := testutil.TestGitlabClient.Projects.GetProjectHook(project, hookID)
		if err != nil {
			return err
		}
		*hook = *gotHook
		return nil
	}
}

func TestResourceGitlabProjectHook_StateUpgradeV0(t *testing.T) {
	t.Parallel()

	givenV0State := map[string]interface{}{
		"project": "foo/bar",
		"hook_id": "42",
		"id":      "42",
	}
	expectedV1State := map[string]interface{}{
		"project": "foo/bar",
		"hook_id": "42",
		"id":      "foo/bar:42",
	}

	actualV1State, err := resourceGitlabProjectHookStateUpgradeV0(context.Background(), givenV0State, nil)
	if err != nil {
		t.Fatalf("Error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expectedV1State, actualV1State) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expectedV1State, actualV1State)
	}
}

type testAccGitlabProjectHookExpectedAttributes struct {
	URL                      string
	PushEvents               bool
	PushEventsBranchFilter   string
	IssuesEvents             bool
	ConfidentialIssuesEvents bool
	MergeRequestsEvents      bool
	TagPushEvents            bool
	NoteEvents               bool
	ConfidentialNoteEvents   bool
	JobEvents                bool
	PipelineEvents           bool
	WikiPageEvents           bool
	DeploymentEvents         bool
	ReleasesEvents           bool
	EnableSSLVerification    bool
}

func testAccCheckGitlabProjectHookAttributes(hook *gitlab.ProjectHook, want *testAccGitlabProjectHookExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if hook.URL != want.URL {
			return fmt.Errorf("got url %q; want %q", hook.URL, want.URL)
		}

		if hook.EnableSSLVerification != want.EnableSSLVerification {
			return fmt.Errorf("got enable_ssl_verification %t; want %t", hook.EnableSSLVerification, want.EnableSSLVerification)
		}

		if hook.PushEvents != want.PushEvents {
			return fmt.Errorf("got push_events %t; want %t", hook.PushEvents, want.PushEvents)
		}

		if hook.PushEventsBranchFilter != want.PushEventsBranchFilter {
			return fmt.Errorf("got push_events_branch_filter %q; want %q", hook.PushEventsBranchFilter, want.PushEventsBranchFilter)
		}

		if hook.IssuesEvents != want.IssuesEvents {
			return fmt.Errorf("got issues_events %t; want %t", hook.IssuesEvents, want.IssuesEvents)
		}

		if hook.ConfidentialIssuesEvents != want.ConfidentialIssuesEvents {
			return fmt.Errorf("got confidential_issues_events %t; want %t", hook.ConfidentialIssuesEvents, want.ConfidentialIssuesEvents)
		}

		if hook.MergeRequestsEvents != want.MergeRequestsEvents {
			return fmt.Errorf("got merge_requests_events %t; want %t", hook.MergeRequestsEvents, want.MergeRequestsEvents)
		}

		if hook.TagPushEvents != want.TagPushEvents {
			return fmt.Errorf("got tag_push_events %t; want %t", hook.TagPushEvents, want.TagPushEvents)
		}

		if hook.NoteEvents != want.NoteEvents {
			return fmt.Errorf("got note_events %t; want %t", hook.NoteEvents, want.NoteEvents)
		}

		if hook.ConfidentialNoteEvents != want.ConfidentialNoteEvents {
			return fmt.Errorf("got confidential_note_events %t; want %t", hook.ConfidentialNoteEvents, want.ConfidentialNoteEvents)
		}

		if hook.JobEvents != want.JobEvents {
			return fmt.Errorf("got job_events %t; want %t", hook.JobEvents, want.JobEvents)
		}

		if hook.PipelineEvents != want.PipelineEvents {
			return fmt.Errorf("got pipeline_events %t; want %t", hook.PipelineEvents, want.PipelineEvents)
		}

		if hook.WikiPageEvents != want.WikiPageEvents {
			return fmt.Errorf("got wiki_page_events %t; want %t", hook.WikiPageEvents, want.WikiPageEvents)
		}

		if hook.DeploymentEvents != want.DeploymentEvents {
			return fmt.Errorf("got deployment_events %t; want %t", hook.DeploymentEvents, want.DeploymentEvents)
		}

		if hook.ReleasesEvents != want.ReleasesEvents {
			return fmt.Errorf("got releases_events %t; want %t", hook.ReleasesEvents, want.ReleasesEvents)
		}

		return nil
	}
}

func testAccCheckGitlabProjectHookDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "gitlab_project_hook" {
			continue
		}

		project, hookID, err := resourceGitlabProjectHookParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, _, err = testutil.TestGitlabClient.Projects.GetProjectHook(project, hookID)
		if err == nil {
			return fmt.Errorf("Project Hook %d in project %s still exists", hookID, project)
		}
		if !api.Is404(err) {
			return err
		}
		return nil
	}
	return nil
}

func testAccGitlabProjectHookConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name = "foo-%d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_project_hook" "foo" {
  project = "${gitlab_project.foo.id}"
  url = "https://example.com/hook-%d"
}
	`, rInt, rInt)
}

func testAccGitlabProjectHookUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name = "foo-%d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_project_hook" "foo" {
  project = "${gitlab_project.foo.id}"
  url = "https://example.com/hook-%d"
  enable_ssl_verification = false
  push_events = true
  push_events_branch_filter = "devel"
  issues_events = false
  confidential_issues_events = false
  merge_requests_events = true
  tag_push_events = true
  note_events = true
  confidential_note_events = true
  job_events = true
  pipeline_events = true
  wiki_page_events = true
  deployment_events = true
  releases_events = true
}
	`, rInt, rInt)
}
