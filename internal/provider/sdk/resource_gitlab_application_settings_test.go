//go:build acceptance
// +build acceptance

package sdk

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGitlabApplicationSettings_basic(t *testing.T) {
	// lintignore:AT001
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		Steps: []resource.TestStep{
			// Verify empty application settings
			{
				Config: `
					resource "gitlab_application_settings" "this" {}
				`,
			},
			// Verify changing some application settings
			{
				Config: `
					resource "gitlab_application_settings" "this" {
						after_sign_up_text = "Welcome to GitLab!"
					}
				`,
			},
		},
	})
}

func TestAccGitlabApplicationSettings_testNullGitProtocol(t *testing.T) {
	// lintignore:AT001
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactoriesV6,
		Steps: []resource.TestStep{
			// Verify we can set the git access to non-nil (which will limit it to just SSH)
			{
				Config: `
					resource "gitlab_application_settings" "this" {
						enabled_git_access_protocol = "ssh"
					}
				`,
				Check: resource.TestCheckResourceAttr("gitlab_application_settings.this", "enabled_git_access_protocol", "ssh"),
			},
			// Verify we can set to "nil" and this works properly.
			{
				Config: `
					resource "gitlab_application_settings" "this" {
						enabled_git_access_protocol = "nil"
					}
				`,
				Check: resource.TestCheckResourceAttr("gitlab_application_settings.this", "enabled_git_access_protocol", ""),
			},
			// Verify we can re-set the git access to non-nil
			{
				Config: `
					resource "gitlab_application_settings" "this" {
						enabled_git_access_protocol = "ssh"
					}
				`,
				Check: resource.TestCheckResourceAttr("gitlab_application_settings.this", "enabled_git_access_protocol", "ssh"),
			},
			// Verify can insensitivity of diffSuppress and the logic
			{
				Config: `
					resource "gitlab_application_settings" "this" {
						enabled_git_access_protocol = "NIL"
					}
				`,
				Check: resource.TestCheckResourceAttr("gitlab_application_settings.this", "enabled_git_access_protocol", ""),
			},
		},
	})
}
