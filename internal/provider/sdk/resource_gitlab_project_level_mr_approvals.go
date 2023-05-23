package sdk

import (
	"context"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/xanzy/go-gitlab"
	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/api"
)

var _ = registerResource("gitlab_project_level_mr_approvals", func() *schema.Resource {
	return &schema.Resource{
		Description: `The ` + "`" + `gitlab_project_level_mr_approval_rule` + "`" + ` resource allows to manage the lifecycle of a Merge Request-level approval rule.

-> This resource requires a GitLab Enterprise instance.

**Upstream API**: [GitLab REST API docs](https://docs.gitlab.com/ee/api/merge_request_approvals.html#merge-request-level-mr-approvals)`,

		CreateContext: resourceGitlabProjectLevelMRApprovalsCreate,
		ReadContext:   resourceGitlabProjectLevelMRApprovalsRead,
		UpdateContext: resourceGitlabProjectLevelMRApprovalsUpdate,
		DeleteContext: resourceGitlabProjectLevelMRApprovalsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema:        gitlabProjectLevelMrApprovalsSchemaV1(),
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceGitlabProjectLevelMrApprovalsResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceGitlabProjectLevelMrApprovalsStateUpgradeV0,
				Version: 0,
			},
		},
	}
})

func gitlabProjectLevelMrApprovalsSchemaV1() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project": {
			Description: "The ID or URL-encoded path of a project to change MR approval configuration.",
			Type:        schema.TypeString,
			ForceNew:    true,
			Required:    true,
		},
		"reset_approvals_on_push": {
			Description: "Set to `true` if you want to remove all approvals in a merge request when new commits are pushed to its source branch. Default is `true`.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"disable_overriding_approvers_per_merge_request": {
			Description: "By default, users are able to edit the approval rules in merge requests. If set to true,",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"merge_requests_author_approval": {
			Description: "Set to `true` if you want to allow merge request authors to self-approve merge requests. Authors",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"merge_requests_disable_committers_approval": {
			Description: "Set to `true` if you want to prevent approval of merge requests by merge request committers.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"require_password_to_approve": {
			Description: "Set to `true` if you want to require authentication when approving a merge request.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
	}
}

func resourceGitlabProjectLevelMrApprovalsResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "The ID of a project to change MR approval configuration.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"reset_approvals_on_push": {
				Description: "Set to `true` if you want to remove all approvals in a merge request when new commits are pushed to its source branch. Default is `true`.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"disable_overriding_approvers_per_merge_request": {
				Description: "By default, users are able to edit the approval rules in merge requests. If set to true,",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"merge_requests_author_approval": {
				Description: "Set to `true` if you want to allow merge request authors to self-approve merge requests. Authors",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"merge_requests_disable_committers_approval": {
				Description: "Set to `true` if you want to prevent approval of merge requests by merge request committers.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"require_password_to_approve": {
				Description: "Set to `true` if you want to require authentication when approving a merge request.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
}

// resourceGitlabProjectLevelMrApprovalsStateUpgradeV0 performs the state migration from V0 to V1.
func resourceGitlabProjectLevelMrApprovalsStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	projectId, ok := rawState["project_id"].(string)
	if !ok {
		projectId = strconv.FormatInt(int64(rawState["project_id"].(float64)), 10)
	}
	rawState["project"] = projectId
	delete(rawState, "project_id")
	tflog.Debug(ctx, "attempting state migration from V0 to V1 - changing the `project_id` attribute to `project`", map[string]interface{}{"project_id": projectId})
	return rawState, nil
}

func resourceGitlabProjectLevelMRApprovalsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)

	project := d.Get("project").(string)

	options := &gitlab.ChangeApprovalConfigurationOptions{
		ResetApprovalsOnPush:                      gitlab.Bool(d.Get("reset_approvals_on_push").(bool)),
		DisableOverridingApproversPerMergeRequest: gitlab.Bool(d.Get("disable_overriding_approvers_per_merge_request").(bool)),
		MergeRequestsAuthorApproval:               gitlab.Bool(d.Get("merge_requests_author_approval").(bool)),
		MergeRequestsDisableCommittersApproval:    gitlab.Bool(d.Get("merge_requests_disable_committers_approval").(bool)),
		RequirePasswordToApprove:                  gitlab.Bool(d.Get("require_password_to_approve").(bool)),
	}

	log.Printf("[DEBUG] Creating new MR approval configuration for project %s:", project)

	if _, _, err := client.Projects.ChangeApprovalConfiguration(project, options, gitlab.WithContext(ctx)); err != nil {
		return diag.Errorf("couldn't create approval configuration: %v", err)
	}

	d.SetId(project)
	return resourceGitlabProjectLevelMRApprovalsRead(ctx, d, meta)
}

func resourceGitlabProjectLevelMRApprovalsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)

	project := d.Id()

	log.Printf("[DEBUG] Reading gitlab approval configuration for project %s", project)

	approvalConfig, _, err := client.Projects.GetApprovalConfiguration(project, gitlab.WithContext(ctx))
	if err != nil {
		if api.Is404(err) {
			log.Printf("[DEBUG] gitlab project approval configuration not found for project %s", project)
			d.SetId("")
			return nil
		}
		return diag.Errorf("couldn't read approval configuration: %v", err)
	}

	d.Set("project", project)
	d.Set("reset_approvals_on_push", approvalConfig.ResetApprovalsOnPush)
	d.Set("disable_overriding_approvers_per_merge_request", approvalConfig.DisableOverridingApproversPerMergeRequest)
	d.Set("merge_requests_author_approval", approvalConfig.MergeRequestsAuthorApproval)
	d.Set("merge_requests_disable_committers_approval", approvalConfig.MergeRequestsDisableCommittersApproval)
	d.Set("require_password_to_approve", approvalConfig.RequirePasswordToApprove)

	return nil
}

func resourceGitlabProjectLevelMRApprovalsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)
	options := &gitlab.ChangeApprovalConfigurationOptions{}

	projectId := d.Id()
	log.Printf("[DEBUG] Updating approval configuration for project %s:", projectId)

	if d.HasChange("reset_approvals_on_push") {
		options.ResetApprovalsOnPush = gitlab.Bool(d.Get("reset_approvals_on_push").(bool))
	}
	if d.HasChange("disable_overriding_approvers_per_merge_request") {
		options.DisableOverridingApproversPerMergeRequest = gitlab.Bool(d.Get("disable_overriding_approvers_per_merge_request").(bool))
	}
	if d.HasChange("merge_requests_author_approval") {
		options.MergeRequestsAuthorApproval = gitlab.Bool(d.Get("merge_requests_author_approval").(bool))
	}
	if d.HasChange("merge_requests_disable_committers_approval") {
		options.MergeRequestsDisableCommittersApproval = gitlab.Bool(d.Get("merge_requests_disable_committers_approval").(bool))
	}
	if d.HasChange("require_password_to_approve") {
		options.RequirePasswordToApprove = gitlab.Bool(d.Get("require_password_to_approve").(bool))
	}

	if _, _, err := client.Projects.ChangeApprovalConfiguration(d.Id(), options, gitlab.WithContext(ctx)); err != nil {
		return diag.Errorf("couldn't update approval configuration: %v", err)
	}

	return resourceGitlabProjectLevelMRApprovalsRead(ctx, d, meta)
}

func resourceGitlabProjectLevelMRApprovalsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)
	projectId := d.Id()

	options := &gitlab.ChangeApprovalConfigurationOptions{
		ResetApprovalsOnPush:                      gitlab.Bool(true),
		DisableOverridingApproversPerMergeRequest: gitlab.Bool(false),
		MergeRequestsAuthorApproval:               gitlab.Bool(false),
		MergeRequestsDisableCommittersApproval:    gitlab.Bool(false),
		RequirePasswordToApprove:                  gitlab.Bool(false),
	}

	log.Printf("[DEBUG] Resetting approval configuration for project %s:", projectId)

	if _, _, err := client.Projects.ChangeApprovalConfiguration(projectId, options, gitlab.WithContext(ctx)); err != nil {
		return diag.Errorf("couldn't reset approval configuration: %v", err)
	}

	return nil
}
