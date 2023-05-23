package sdk

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/xanzy/go-gitlab"
	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/api"
	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/utils"
)

var _ = registerResource("gitlab_project_access_token", func() *schema.Resource {
	return &schema.Resource{
		Description: `The ` + "`" + `gitlab_project_access_token` + "`" + ` resource allows to manage the lifecycle of a project access token.

**Upstream API**: [GitLab API docs](https://docs.gitlab.com/ee/api/project_access_tokens.html)`,

		CreateContext: resourceGitlabProjectAccessTokenCreate,
		ReadContext:   resourceGitlabProjectAccessTokenRead,
		DeleteContext: resourceGitlabProjectAccessTokenDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"project": {
				Description: "The id of the project to add the project access token to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "A name to describe the project access token.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"scopes": {
				Description: "Valid values: `api`, `read_api`, `read_repository`, `write_repository`, `read_registry`, `write_registry`.",
				Type:        schema.TypeSet,
				Required:    true,
				ForceNew:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"api", "read_api", "read_repository", "write_repository", "read_registry", "write_registry"}, false),
				},
			},
			"expires_at": {
				Description:      "Time the token will expire it, YYYY-MM-DD format.",
				Type:             schema.TypeString,
				ValidateDiagFunc: isISO6801Date,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
			},
			"token": {
				Description: "The secret token. **Note**: the token is not available for imported resources.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"active": {
				Description: "True if the token is active.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"created_at": {
				Description: "Time the token has been created, RFC3339 format.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"revoked": {
				Description: "True if the token is revoked.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"user_id": {
				Description: "The user_id associated to the token.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"access_level": {
				Description:      fmt.Sprintf("The access level for the project access token. Valid values are: %s. Default is `%s`.", utils.RenderValueListForDocs(api.ValidProjectAccessLevelNames), api.AccessLevelValueToName[gitlab.MaintainerPermissions]),
				Type:             schema.TypeString,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(api.ValidProjectAccessLevelNames, false)),
				Optional:         true,
				Default:          api.AccessLevelValueToName[gitlab.MaintainerPermissions],
				ForceNew:         true,
			},
		},
	}
})

func resourceGitlabProjectAccessTokenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)
	accessLevelId := api.AccessLevelNameToValue[d.Get("access_level").(string)]
	project := d.Get("project").(string)

	options := &gitlab.CreateProjectAccessTokenOptions{
		Name:        gitlab.String(d.Get("name").(string)),
		Scopes:      stringSetToStringSlice(d.Get("scopes").(*schema.Set)),
		AccessLevel: &accessLevelId,
	}

	log.Printf("[DEBUG] create gitlab ProjectAccessToken %s %s for project ID %s", *options.Name, options.Scopes, project)

	if v, ok := d.GetOk("expires_at"); ok {
		parsedExpiresAt, err := time.Parse("2006-01-02", v.(string))
		if err != nil {
			return diag.Errorf("Invalid expires_at date: %v", err)
		}
		parsedExpiresAtISOTime := gitlab.ISOTime(parsedExpiresAt)
		options.ExpiresAt = &parsedExpiresAtISOTime
	}

	projectAccessToken, _, err := client.ProjectAccessTokens.CreateProjectAccessToken(project, options, gitlab.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	PATstring := strconv.Itoa(projectAccessToken.ID)
	d.SetId(utils.BuildTwoPartID(&project, &PATstring))
	d.Set("token", projectAccessToken.Token)

	return resourceGitlabProjectAccessTokenRead(ctx, d, meta)
}

func resourceGitlabProjectAccessTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	project, PATstring, err := utils.ParseTwoPartID(d.Id())
	if err != nil {
		return diag.Errorf("Error parsing ID: %s", d.Id())
	}

	client := meta.(*gitlab.Client)

	projectAccessTokenID, err := strconv.Atoi(PATstring)
	if err != nil {
		return diag.Errorf("%s cannot be converted to int", PATstring)
	}

	log.Printf("[DEBUG] read gitlab ProjectAccessToken %d, project ID %s", projectAccessTokenID, project)

	projectAccessToken, _, err := client.ProjectAccessTokens.GetProjectAccessToken(project, projectAccessTokenID, gitlab.WithContext(ctx))
	if err != nil {
		if api.Is404(err) {
			log.Printf("[DEBUG] GitLab ProjectAccessToken %d, project ID %s not found, removing from state", projectAccessTokenID, project)
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("project", project)
	d.Set("name", projectAccessToken.Name)
	if projectAccessToken.ExpiresAt != nil {
		d.Set("expires_at", projectAccessToken.ExpiresAt.String())
	}
	d.Set("active", projectAccessToken.Active)
	d.Set("created_at", projectAccessToken.CreatedAt.String())
	d.Set("revoked", projectAccessToken.Revoked)
	d.Set("user_id", projectAccessToken.UserID)
	d.Set("access_level", api.AccessLevelValueToName[projectAccessToken.AccessLevel])
	if err = d.Set("scopes", projectAccessToken.Scopes); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceGitlabProjectAccessTokenDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	project, patString, err := utils.ParseTwoPartID(d.Id())
	if err != nil {
		return diag.Errorf("Error parsing ID: %s", d.Id())
	}

	client := meta.(*gitlab.Client)

	projectAccessTokenID, err := strconv.Atoi(patString)
	if err != nil {
		return diag.Errorf("%s cannot be converted to int", patString)
	}

	log.Printf("[DEBUG] Delete gitlab ProjectAccessToken %s", d.Id())
	_, err = client.ProjectAccessTokens.RevokeProjectAccessToken(project, projectAccessTokenID, gitlab.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Waiting for ProjectAccessToken %s to finish deleting", d.Id())

	err = retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
		_, _, err := client.ProjectAccessTokens.GetProjectAccessToken(project, projectAccessTokenID, gitlab.WithContext(ctx))
		if err != nil {
			if api.Is404(err) {
				return nil
			}
			return retry.NonRetryableError(err)
		}
		return retry.RetryableError(errors.New("project access token was not deleted"))
	})

	return diag.FromErr(err)
}
