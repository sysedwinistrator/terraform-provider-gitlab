package sdk

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/xanzy/go-gitlab"
	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/api"
	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/utils"
)

var _ = registerResource("gitlab_project_membership", func() *schema.Resource {
	return &schema.Resource{
		Description: `The ` + "`gitlab_project_membership`" + ` resource allows to manage the lifecycle of a users project membership.

-> If a project should grant membership to an entire group use the ` + "`gitlab_project_share_group`" + ` resource instead.

**Upstream API**: [GitLab REST API docs](https://docs.gitlab.com/ee/api/members.html)`,

		CreateContext: resourceGitlabProjectMembershipCreate,
		ReadContext:   resourceGitlabProjectMembershipRead,
		UpdateContext: resourceGitlabProjectMembershipUpdate,
		DeleteContext: resourceGitlabProjectMembershipDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"project": {
				Description: "The ID or URL-encoded path of the project.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"user_id": {
				Description: "The id of the user.",
				Type:        schema.TypeInt,
				ForceNew:    true,
				Required:    true,
			},
			"access_level": {
				Description:      fmt.Sprintf("The access level for the member. Valid values are: %s", utils.RenderValueListForDocs(api.ValidProjectAccessLevelNames)),
				Type:             schema.TypeString,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(api.ValidProjectAccessLevelNames, false)),
				Required:         true,
			},
			"expires_at": {
				Description:  "Expiration date for the project membership. Format: `YYYY-MM-DD`",
				Type:         schema.TypeString,
				ValidateFunc: validateDateFunc,
				Optional:     true,
			},
		},
	}
})

func resourceGitlabProjectMembershipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)

	userId := d.Get("user_id").(int)
	project := d.Get("project").(string)
	expiresAt := d.Get("expires_at").(string)
	accessLevelId := api.AccessLevelNameToValue[d.Get("access_level").(string)]

	options := &gitlab.AddProjectMemberOptions{
		UserID:      &userId,
		AccessLevel: &accessLevelId,
		ExpiresAt:   &expiresAt,
	}
	log.Printf("[DEBUG] create gitlab project membership for %d in %s", options.UserID, project)

	_, _, err := client.ProjectMembers.AddProjectMember(project, options, gitlab.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	userIdString := strconv.Itoa(userId)
	d.SetId(utils.BuildTwoPartID(&project, &userIdString))
	return resourceGitlabProjectMembershipRead(ctx, d, meta)
}

func resourceGitlabProjectMembershipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)
	id := d.Id()
	log.Printf("[DEBUG] read gitlab project projectMember %s", id)

	project, userId, err := projectAndUserIdFromId(id)
	if err != nil {
		return diag.FromErr(err)
	}

	projectMember, _, err := client.ProjectMembers.GetProjectMember(project, userId, gitlab.WithContext(ctx))
	if err != nil {
		if api.Is404(err) {
			log.Printf("[DEBUG] gitlab project membership for %s not found so removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	resourceGitlabProjectMembershipSetToState(d, projectMember, &project)
	return nil
}

func projectAndUserIdFromId(id string) (string, int, error) {
	project, userIdString, err := utils.ParseTwoPartID(id)
	userId, e := strconv.Atoi(userIdString)
	if err != nil {
		e = err
	}
	if e != nil {
		log.Printf("[WARN] cannot get project member id from input: %v", id)
	}
	return project, userId, e
}

func resourceGitlabProjectMembershipUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)

	userId := d.Get("user_id").(int)
	project := d.Get("project").(string)
	expiresAt := d.Get("expires_at").(string)
	accessLevelId := api.AccessLevelNameToValue[strings.ToLower(d.Get("access_level").(string))]

	options := gitlab.EditProjectMemberOptions{
		AccessLevel: &accessLevelId,
		ExpiresAt:   &expiresAt,
	}
	log.Printf("[DEBUG] update gitlab project membership %v for %s", userId, project)

	_, _, err := client.ProjectMembers.EditProjectMember(project, userId, &options, gitlab.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceGitlabProjectMembershipRead(ctx, d, meta)
}

func resourceGitlabProjectMembershipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)

	id := d.Id()
	project, userId, err := projectAndUserIdFromId(id)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Delete gitlab project membership %v for %s", userId, project)

	_, err = client.ProjectMembers.DeleteProjectMember(project, userId, gitlab.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceGitlabProjectMembershipSetToState(d *schema.ResourceData, projectMember *gitlab.ProjectMember, projectId *string) {

	d.Set("project", projectId)
	d.Set("user_id", projectMember.ID)
	d.Set("access_level", api.AccessLevelValueToName[projectMember.AccessLevel])
	if projectMember.ExpiresAt != nil {
		d.Set("expires_at", projectMember.ExpiresAt.String())
	} else {
		d.Set("expires_at", "")
	}
	userId := strconv.Itoa(projectMember.ID)
	d.SetId(utils.BuildTwoPartID(projectId, &userId))
}
