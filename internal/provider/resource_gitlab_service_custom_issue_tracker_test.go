//go:build acceptance
// +build acceptance

package provider

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/testutil"
	"regexp"
	"testing"
)

func TestAcc_GitlabServiceCustomIssueTracker_basic(t *testing.T) {
	testProject := testutil.CreateProject(t)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAcc_GitlabServiceCustomIssueTracker_CheckDestroy(testProject.ID),
		Steps: []resource.TestStep{
			// Create a Custom Issue Tracker service
			{
				Config: fmt.Sprintf(`
				resource "gitlab_service_custom_issue_tracker" "this" {
					project     = "%s"
					project_url = "https://customtracker.com"
					issues_url  = "https://customtracker.com/:id"
				}
				`, testProject.PathWithNamespace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("gitlab_service_custom_issue_tracker.this", "id"),
					resource.TestCheckResourceAttrSet("gitlab_service_custom_issue_tracker.this", "project"),
					resource.TestCheckResourceAttrSet("gitlab_service_custom_issue_tracker.this", "project_url"),
					resource.TestCheckResourceAttrSet("gitlab_service_custom_issue_tracker.this", "issues_url"),
					resource.TestCheckResourceAttrSet("gitlab_service_custom_issue_tracker.this", "active"),
					resource.TestCheckResourceAttrSet("gitlab_service_custom_issue_tracker.this", "created_at"),
				),
			},
			// Verify upstream attributes with an import.
			{
				ResourceName:      "gitlab_service_custom_issue_tracker.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update the Custom Issue Tracker service
			{
				Config: fmt.Sprintf(`
				resource "gitlab_service_custom_issue_tracker" "this" {
					project     = %d
					project_url = "https://anotherracker.com"
					issues_url  = "https://anotherracker.com/:id"
				}
				`, testProject.ID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("gitlab_service_custom_issue_tracker.this", "id"),
					resource.TestCheckResourceAttrSet("gitlab_service_custom_issue_tracker.this", "project"),
					resource.TestCheckResourceAttrSet("gitlab_service_custom_issue_tracker.this", "project_url"),
					resource.TestCheckResourceAttrSet("gitlab_service_custom_issue_tracker.this", "issues_url"),
					resource.TestCheckResourceAttrSet("gitlab_service_custom_issue_tracker.this", "active"),
					resource.TestCheckResourceAttrSet("gitlab_service_custom_issue_tracker.this", "created_at"),
					resource.TestCheckResourceAttrSet("gitlab_service_custom_issue_tracker.this", "updated_at"),
				),
			},
			// Verify upstream attributes with an import.
			{
				ResourceName:      "gitlab_service_custom_issue_tracker.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_GitlabServiceCustomIssueTracker_failures(t *testing.T) {
	testProject := testutil.CreateProject(t)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAcc_GitlabServiceCustomIssueTracker_CheckDestroy(testProject.ID),
		Steps: []resource.TestStep{
			// Fail if project missing
			{
				Config: `
             		resource "gitlab_service_custom_issue_tracker" "this" {
						project_url = "https://customtracker.org"
						issues_url  = "https://customtracker.org/:id"
					}`,
				ExpectError: regexp.MustCompile(`The argument "project" is required, but no definition was found`),
			},
			// Fail if project_url missing
			{
				Config: fmt.Sprintf(`
				resource "gitlab_service_custom_issue_tracker" "this" {
					project    = %d
					issues_url = "https://customtracker.org/:id"
				}`, testProject.ID),
				ExpectError: regexp.MustCompile(`The argument "project_url" is required, but no definition was found`),
			},
			// Fail if project_url is invalid
			{
				Config: fmt.Sprintf(`
				resource "gitlab_service_custom_issue_tracker" "this" {
					project     = %d
					project_url = "customtracker.org"
					issues_url  = "https://customtracker.org/:id"
				}`, testProject.ID),
				ExpectError: regexp.MustCompile(`Attribute project_url value should be an URL with http or https schema`),
			},
			// Fail if issues_url missing
			{
				Config: fmt.Sprintf(`
				resource "gitlab_service_custom_issue_tracker" "this" {
					project     = %d
					project_url = "https://customtracker.org"
				}`, testProject.ID),
				ExpectError: regexp.MustCompile(`The argument "issues_url" is required, but no definition was found`),
			},
			// Fail if issues_url is invalid
			{
				Config: fmt.Sprintf(`
				resource "gitlab_service_custom_issue_tracker" "this" {
					project     = %d
					project_url = "https://customtracker.org"
					issues_url  = "customtracker.org/:id"
				}`, testProject.ID),
				ExpectError: regexp.MustCompile(`Attribute issues_url value should be an URL with http or https schema`),
			},
			// Fail if issues_url doesn't contain :id
			{
				Config: fmt.Sprintf(`
				resource "gitlab_service_custom_issue_tracker" "this" {
					project     = %d
					project_url = "https://customtracker.org"
					issues_url  = "https://customtracker.org/no-id"
				}`, testProject.ID),
				ExpectError: regexp.MustCompile(`Attribute issues_url value should contain :id placeholder`),
			},
		},
	})
}

func testAcc_GitlabServiceCustomIssueTracker_CheckDestroy(projectId int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		service, _, err := testutil.TestGitlabClient.Services.GetCustomIssueTrackerService(projectId)
		if err != nil {
			return fmt.Errorf("Error calling API to get the Custom Issue Tracker: %w", err)
		}
		if service != nil && service.Active != false {
			return errors.New("Custom issue tracker still exists")
		}
		return nil
	}
}
