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

func TestAccGitlabPipelineSchedule_StateUpgradeV0(t *testing.T) {
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
			actualV1State, err := resourceGitlabPipelineScheduleStateUpgradeV0(context.Background(), tc.givenV0State, nil)
			if err != nil {
				t.Fatalf("Error migrating state: %s", err)
			}

			if !reflect.DeepEqual(tc.expectedV1State, actualV1State) {
				t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", tc.expectedV1State, actualV1State)
			}
		})

	}
}

func TestAccGitlabPipelineSchedule_basic(t *testing.T) {
	var schedule gitlab.PipelineSchedule
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		CheckDestroy:             testAccCheckGitlabPipelineScheduleDestroy,
		Steps: []resource.TestStep{
			// Create a project and pipeline schedule with default options
			{
				Config: testAccGitlabPipelineScheduleConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabPipelineScheduleExists("gitlab_pipeline_schedule.schedule", &schedule),
					testAccCheckGitlabPipelineScheduleAttributes(&schedule, &testAccGitlabPipelineScheduleExpectedAttributes{
						Description:  "Pipeline Schedule",
						Ref:          "master",
						Cron:         "0 1 * * *",
						CronTimezone: "UTC",
						Active:       true,
					}),
				),
			},
			// Verify Import
			{
				ResourceName:      "gitlab_pipeline_schedule.schedule",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update the pipeline schedule to change the parameters
			{
				Config: testAccGitlabPipelineScheduleUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabPipelineScheduleExists("gitlab_pipeline_schedule.schedule", &schedule),
					testAccCheckGitlabPipelineScheduleAttributes(&schedule, &testAccGitlabPipelineScheduleExpectedAttributes{
						Description:  "Schedule",
						Ref:          "master",
						Cron:         "0 4 * * *",
						CronTimezone: "UTC",
						Active:       false,
					}),
				),
			},
			// Verify Import
			{
				ResourceName:      "gitlab_pipeline_schedule.schedule",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update the pipeline schedule to get back to initial settings
			{
				Config: testAccGitlabPipelineScheduleConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabPipelineScheduleExists("gitlab_pipeline_schedule.schedule", &schedule),
					testAccCheckGitlabPipelineScheduleAttributes(&schedule, &testAccGitlabPipelineScheduleExpectedAttributes{
						Description:  "Pipeline Schedule",
						Ref:          "master",
						Cron:         "0 1 * * *",
						CronTimezone: "UTC",
						Active:       true,
					}),
				),
			},
			// Verify Import
			{
				ResourceName:      "gitlab_pipeline_schedule.schedule",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckGitlabPipelineScheduleExists(n string, schedule *gitlab.PipelineSchedule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		project, pipelineScheduleId, err := resourceGitlabPipelineTriggerParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		sc, _, err := testutil.TestGitlabClient.PipelineSchedules.GetPipelineSchedule(project, pipelineScheduleId)
		if err != nil {
			if api.Is404(err) {
				return fmt.Errorf("Pipeline Schedule %q does not exist", rs.Primary.ID)
			}
			return err
		}
		*schedule = *sc
		return nil
	}
}

type testAccGitlabPipelineScheduleExpectedAttributes struct {
	Description  string
	Ref          string
	Cron         string
	CronTimezone string
	Active       bool
}

func testAccCheckGitlabPipelineScheduleAttributes(schedule *gitlab.PipelineSchedule, want *testAccGitlabPipelineScheduleExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if schedule.Description != want.Description {
			return fmt.Errorf("got description %q; want %q", schedule.Description, want.Description)
		}
		if schedule.Ref != want.Ref {
			return fmt.Errorf("got ref %q; want %q", schedule.Ref, want.Ref)
		}

		if schedule.Cron != want.Cron {
			return fmt.Errorf("got cron %q; want %q", schedule.Cron, want.Cron)
		}

		if schedule.CronTimezone != want.CronTimezone {
			return fmt.Errorf("got cron_timezone %q; want %q", schedule.CronTimezone, want.CronTimezone)
		}

		if schedule.Active != want.Active {
			return fmt.Errorf("got active %t; want %t", schedule.Active, want.Active)
		}

		return nil
	}
}

func testAccCheckGitlabPipelineScheduleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "gitlab_pipeline_schedule" {
			continue
		}

		project, pipelineScheduleId, err := resourceGitlabPipelineTriggerParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, _, err = testutil.TestGitlabClient.PipelineSchedules.GetPipelineSchedule(project, pipelineScheduleId)
		if err == nil {
			return fmt.Errorf("the Pipeline Schedule %d in project %s still exists", pipelineScheduleId, project)
		}
		if !api.Is404(err) {
			return err
		}
		return nil
	}
	return nil
}

func testAccGitlabPipelineScheduleConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name = "foo-%d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_pipeline_schedule" "schedule" {
	project = "${gitlab_project.foo.id}"
	description = "Pipeline Schedule"
	ref = "master"
	cron = "0 1 * * *"
}
	`, rInt)
}

func testAccGitlabPipelineScheduleUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "gitlab_project" "foo" {
  name = "foo-%d"
  description = "Terraform acceptance tests"

  # So that acceptance tests can be run in a gitlab organization
  # with no billing
  visibility_level = "public"
}

resource "gitlab_pipeline_schedule" "schedule" {
  project = "${gitlab_project.foo.id}"
  description = "Schedule"
  ref = "master"
  cron = "0 4 * * *"
  active = false
}
	`, rInt)
}
