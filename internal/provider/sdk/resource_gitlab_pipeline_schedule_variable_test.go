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

	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/testutil"
)

func TestAccGitlabPipelineScheduleVariable_StateUpgradeV0(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name            string
		givenV0State    map[string]interface{}
		expectedV1State map[string]interface{}
	}{
		{
			name: "Project With ID",
			givenV0State: map[string]interface{}{
				"project":              "99",
				"pipeline_schedule_id": 42,
				"key":                  "some-key",
				"id":                   "42:some-key",
			},
			expectedV1State: map[string]interface{}{
				"project":              "99",
				"pipeline_schedule_id": 42,
				"key":                  "some-key",
				"id":                   "99:42:some-key",
			},
		},
		{
			name: "Project With ID and pipeline schedule id as float",
			givenV0State: map[string]interface{}{
				"project":              "99",
				"pipeline_schedule_id": 42.0,
				"key":                  "some-key",
				"id":                   "42:some-key",
			},
			expectedV1State: map[string]interface{}{
				"project":              "99",
				"pipeline_schedule_id": 42.0,
				"key":                  "some-key",
				"id":                   "99:42:some-key",
			},
		},
		{
			name: "Project With Namespace",
			givenV0State: map[string]interface{}{
				"project":              "foo/bar",
				"pipeline_schedule_id": 42,
				"key":                  "some-key",
				"id":                   "42:some-key",
			},
			expectedV1State: map[string]interface{}{
				"project":              "foo/bar",
				"pipeline_schedule_id": 42,
				"key":                  "some-key",
				"id":                   "foo/bar:42:some-key",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualV1State, err := resourceGitlabProjectLabelStateUpgradeV0(context.Background(), tc.givenV0State, nil)
			if err != nil {
				t.Fatalf("Error migrating state: %s", err)
			}

			if !reflect.DeepEqual(tc.expectedV1State, actualV1State) {
				t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", tc.expectedV1State, actualV1State)
			}
		})

	}
}

func TestAccGitlabPipelineScheduleVariable_basic(t *testing.T) {
	var variable gitlab.PipelineVariable
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		CheckDestroy:             testAccCheckGitlabPipelineScheduleVariableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGitlabPipelineScheduleVariableConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabPipelineScheduleVariableExists("gitlab_pipeline_schedule_variable.schedule_var", &variable),
					testAccCheckGitlabPipelineScheduleVariableAttributes(&variable, &testAccGitlabPipelineScheduleVariableExpectedAttributes{
						Key:   "TERRAFORMED_TEST_VALUE",
						Value: "test",
					}),
				),
			},
			// Verify Import
			{
				ResourceName:      "gitlab_pipeline_schedule_variable.schedule_var",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGitlabPipelineScheduleVariableUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabPipelineScheduleVariableExists("gitlab_pipeline_schedule_variable.schedule_var", &variable),
					testAccCheckGitlabPipelineScheduleVariableAttributes(&variable, &testAccGitlabPipelineScheduleVariableExpectedAttributes{
						Key:   "TERRAFORMED_TEST_VALUE",
						Value: "test_updated",
					}),
				),
			},
			// Verify Import
			{
				ResourceName:      "gitlab_pipeline_schedule_variable.schedule_var",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGitlabPipelineScheduleVariableConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabPipelineScheduleVariableExists("gitlab_pipeline_schedule_variable.schedule_var", &variable),
					testAccCheckGitlabPipelineScheduleVariableAttributes(&variable, &testAccGitlabPipelineScheduleVariableExpectedAttributes{
						Key:   "TERRAFORMED_TEST_VALUE",
						Value: "test",
					}),
				),
			},
			// Verify Import
			{
				ResourceName:      "gitlab_pipeline_schedule_variable.schedule_var",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckGitlabPipelineScheduleVariableExists(n string, variable *gitlab.PipelineVariable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		project, scheduleID, variableKey, err := resourceGitlabPipelineScheduleVariableParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		pipelineSchedule, _, err := testutil.TestGitlabClient.PipelineSchedules.GetPipelineSchedule(project, scheduleID)
		if err != nil {
			return err
		}

		for _, pipelineVariable := range pipelineSchedule.Variables {
			if pipelineVariable.Key == variableKey {
				*variable = *pipelineVariable
				return nil
			}
		}
		return fmt.Errorf("PipelineScheduleVariable %s does not exist", variable.Key)
	}
}

type testAccGitlabPipelineScheduleVariableExpectedAttributes struct {
	Key   string
	Value string
}

func testAccCheckGitlabPipelineScheduleVariableAttributes(variable *gitlab.PipelineVariable, want *testAccGitlabPipelineScheduleVariableExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if variable.Key != want.Key {
			return fmt.Errorf("got key %q; want %q", variable.Key, want.Key)
		}

		return nil
	}
}

func testAccGitlabPipelineScheduleVariableConfig(rInt int) string {
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

resource "gitlab_pipeline_schedule_variable" "schedule_var" {
	project = "${gitlab_project.foo.id}"
	pipeline_schedule_id = "${gitlab_pipeline_schedule.schedule.pipeline_schedule_id}"
	key = "TERRAFORMED_TEST_VALUE"
	value = "test"
}
	`, rInt)
}

func testAccGitlabPipelineScheduleVariableUpdateConfig(rInt int) string {
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

resource "gitlab_pipeline_schedule_variable" "schedule_var" {
	project = "${gitlab_project.foo.id}"
	pipeline_schedule_id = "${gitlab_pipeline_schedule.schedule.pipeline_schedule_id}"
	key = "TERRAFORMED_TEST_VALUE"
	value = "test_updated"
}
	`, rInt)
}

func testAccCheckGitlabPipelineScheduleVariableDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "gitlab_pipeline_schedule_variable" {
			continue
		}

		project, scheduleID, variableKey, err := resourceGitlabPipelineScheduleVariableParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		gotPS, _, err := testutil.TestGitlabClient.PipelineSchedules.GetPipelineSchedule(project, scheduleID)
		if err == nil {
			for _, v := range gotPS.Variables {
				if v.Key == variableKey {
					return fmt.Errorf("pipeline schedule variable still exists")
				}
			}
		}
	}
	return nil
}
