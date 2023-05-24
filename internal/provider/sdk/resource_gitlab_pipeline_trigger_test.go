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

func TestAccGitlabPipelineTrigger_StateUpgradeV0(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name            string
		givenV0State    map[string]interface{}
		expectedV1State map[string]interface{}
	}{
		{
			name: "Project With ID",
			givenV0State: map[string]interface{}{
				"project": "99",
				"id":      "42",
			},
			expectedV1State: map[string]interface{}{
				"project": "99",
				"id":      "99:42",
			},
		},
		{
			name: "Project With Namespace",
			givenV0State: map[string]interface{}{
				"project": "foo/bar",
				"id":      "42",
			},
			expectedV1State: map[string]interface{}{
				"project": "foo/bar",
				"id":      "foo/bar:42",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualV1State, err := resourceGitlabPipelineTriggerStateUpgradeV0(context.Background(), tc.givenV0State, nil)
			if err != nil {
				t.Fatalf("Error migrating state: %s", err)
			}

			if !reflect.DeepEqual(tc.expectedV1State, actualV1State) {
				t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", tc.expectedV1State, actualV1State)
			}
		})

	}
}

func TestAccGitlabPipelineTrigger_SchemaMigration0_1(t *testing.T) {
	testProject := testutil.CreateProject(t)

	config := fmt.Sprintf(`
	resource "gitlab_pipeline_trigger" "trigger" {
		project = "%d"
		description = "External Pipeline Trigger"
	}
		`, testProject.ID)

	resource.ParallelTest(t, resource.TestCase{
		CheckDestroy: testAccCheckGitlabPipelineTriggerDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"gitlab": {
						VersionConstraint: "~> 15.7.0", // Earliest 15.X deployment
						Source:            "gitlabhq/gitlab",
					},
				},
				Config: config,
			},
			{
				ProtoV6ProviderFactories: providerFactoriesV6,
				Config:                   config,
				PlanOnly:                 true,
			},
		},
	})
}

func TestAccGitlabPipelineTrigger_basic(t *testing.T) {
	var trigger gitlab.PipelineTrigger
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		CheckDestroy:             testAccCheckGitlabPipelineTriggerDestroy,
		Steps: []resource.TestStep{
			// Create a project and pipeline trigger with default options
			{
				Config: testAccGitlabPipelineTriggerConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabPipelineTriggerExists("gitlab_pipeline_trigger.trigger", &trigger),
					testAccCheckGitlabPipelineTriggerAttributes(&trigger, &testAccGitlabPipelineTriggerExpectedAttributes{
						Description: "External Pipeline Trigger",
					}),
				),
			},
			// Verify Import
			{
				ResourceName:      "gitlab_pipeline_trigger.trigger",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update the pipeline trigger to change the parameters
			{
				Config: testAccGitlabPipelineTriggerUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabPipelineTriggerExists("gitlab_pipeline_trigger.trigger", &trigger),
					testAccCheckGitlabPipelineTriggerAttributes(&trigger, &testAccGitlabPipelineTriggerExpectedAttributes{
						Description: "Trigger",
					}),
				),
			},
			// Verify Import
			{
				ResourceName:      "gitlab_pipeline_trigger.trigger",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update the pipeline trigger to get back to initial settings
			{
				Config: testAccGitlabPipelineTriggerConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabPipelineTriggerExists("gitlab_pipeline_trigger.trigger", &trigger),
					testAccCheckGitlabPipelineTriggerAttributes(&trigger, &testAccGitlabPipelineTriggerExpectedAttributes{
						Description: "External Pipeline Trigger",
					}),
				),
			},
			// Verify Import
			{
				ResourceName:      "gitlab_pipeline_trigger.trigger",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckGitlabPipelineTriggerExists(n string, trigger *gitlab.PipelineTrigger) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		project, pipelineTriggerId, err := resourceGitlabPipelineTriggerParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		t, _, err := testutil.TestGitlabClient.PipelineTriggers.GetPipelineTrigger(project, pipelineTriggerId)
		if err != nil {
			if api.Is404(err) {
				return fmt.Errorf("Pipeline Trigger %q does not exist", rs.Primary.ID)
			}
			return err
		}
		*trigger = *t
		return nil
	}
}

type testAccGitlabPipelineTriggerExpectedAttributes struct {
	Description string
}

func testAccCheckGitlabPipelineTriggerAttributes(trigger *gitlab.PipelineTrigger, want *testAccGitlabPipelineTriggerExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if trigger.Description != want.Description {
			return fmt.Errorf("got description %q; want %q", trigger.Description, want.Description)
		}

		return nil
	}
}

func testAccCheckGitlabPipelineTriggerDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "gitlab_pipeline_trigger" {
			continue
		}

		project, pipelineTriggerId, err := resourceGitlabPipelineTriggerParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, _, err = testutil.TestGitlabClient.PipelineTriggers.GetPipelineTrigger(project, pipelineTriggerId)
		if err == nil {
			return fmt.Errorf("the Pipeline Trigger %d in project %s still exists", pipelineTriggerId, project)
		}
		if !api.Is404(err) {
			return err
		}
		return nil
	}
	return nil
}

func testAccGitlabPipelineTriggerConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name = "foo-%d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_pipeline_trigger" "trigger" {
	project = "${gitlab_project.foo.id}"
	description = "External Pipeline Trigger"
}
	`, rInt)
}

func testAccGitlabPipelineTriggerUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name = "foo-%d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_pipeline_trigger" "trigger" {
  project = "${gitlab_project.foo.id}"
  description = "Trigger"
}
	`, rInt)
}
