package sdk

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/xanzy/go-gitlab"
	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/api"
	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/utils"
)

var _ = registerResource("gitlab_group_ldap_link", func() *schema.Resource {
	return &schema.Resource{
		Description: `The ` + "`gitlab_group_ldap_link`" + ` resource allows to manage the lifecycle of an LDAP integration with a group.

**Upstream API**: [GitLab REST API docs](https://docs.gitlab.com/ee/api/groups.html#ldap-group-links)`,

		CreateContext: resourceGitlabGroupLdapLinkCreate,
		ReadContext:   resourceGitlabGroupLdapLinkRead,
		DeleteContext: resourceGitlabGroupLdapLinkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema:        gitlabGroupLDAPLinkSchema(),
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceGitlabGroupLDAPLinkResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceGitlabGroupLDAPLinkStateUpgradeV0,
				Version: 0,
			},
		},
	}
})

func gitlabGroupLDAPLinkSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"group": {
			Description: "The ID or URL-encoded path of the group",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"cn": {
			Description:   "The CN of the LDAP group to link with. Required if `filter` is not provided.",
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"filter"},
		},
		"filter": {
			Description:   "The LDAP filter for the group. Required if `cn` is not provided. Requires GitLab Premium or above.",
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"cn"},
		},
		"access_level": {
			Description:      fmt.Sprintf("Minimum access level for members of the LDAP group. Valid values are: %s", utils.RenderValueListForDocs(api.ValidGroupAccessLevelNames)),
			Type:             schema.TypeString,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(api.ValidGroupAccessLevelNames, false)),
			Optional:         true,
			ForceNew:         true,
			Deprecated:       "Use `group_access` instead of the `access_level` attribute.",
			ExactlyOneOf:     []string{"access_level", "group_access"},
		},
		"group_access": {
			Description:      fmt.Sprintf("Minimum access level for members of the LDAP group. Valid values are: %s", utils.RenderValueListForDocs(api.ValidGroupAccessLevelNames)),
			Type:             schema.TypeString,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(api.ValidGroupAccessLevelNames, false)),
			Optional:         true,
			ForceNew:         true,
			ExactlyOneOf:     []string{"access_level", "group_access"},
		},
		// Changing GitLab API parameter "provider" to "ldap_provider" to avoid clashing with the Terraform "provider" key word
		"ldap_provider": {
			Description: "The name of the LDAP provider as stored in the GitLab database. Note that this is NOT the value of the `label` attribute as shown in the web UI. In most cases this will be `ldapmain` but you may use the [LDAP check rake task](https://docs.gitlab.com/ee/administration/raketasks/ldap.html#check) for receiving the LDAP server name: `LDAP: ... Server: ldapmain`",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"force": {
			Description: "If true, then delete and replace an existing LDAP link if one exists.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			ForceNew:    true,
		},
	}
}

// resourceGitlabGroupLDAPLinkResourceV0 returns the V0 schema definition.
// From V0-V1 the `id` attribute value format changed from `<ldap_provider>:<cn>` to `<group>:<ldap_provider>:<cn>:<filter>`,
// which means that the actual schema definition was not impacted and we can just return the
// V1 schema as V0 schema.
func resourceGitlabGroupLDAPLinkResourceV0() *schema.Resource {
	return &schema.Resource{Schema: gitlabGroupLDAPLinkSchema()}
}

// resourceGitlabProjectLabelStateUpgradeV0 performs the state migration from V0 to V1.
func resourceGitlabGroupLDAPLinkStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	group := ""
	// check to determine if "group_id" is present. If it is, use that, otherwise use "group". This is because
	// "group_id" changed to "group" in 16.0, so the previous state may use either.
	if rawState["group_id"] != nil {
		group = rawState["group_id"].(string)
	} else {
		group = rawState["group"].(string)
	}

	ldap := rawState["ldap_provider"].(string)
	cn := rawState["cn"].(string)

	// Filter was not a supported attribute prior to 16.0 where this migration was added, so it will always be empty
	// However, we'll handle it here _just in case_.
	filter := rawState["filter"].(string)

	oldId := rawState["id"].(string)
	tflog.Debug(ctx, "attempting state migration from V0 to V1 - changing the `id` attribute format to include the group", map[string]interface{}{"group_id": group, "v0-id": oldId})
	rawState["id"] = resourceGitLabGroupLDAPLinkBuildId(group, ldap, cn, filter)

	tflog.Debug(ctx, "migrated `id` attribute for V0 to V1", map[string]interface{}{"v0-id": oldId, "v1-id": rawState["id"]})
	return rawState, nil
}

// Builds the 3-part ID for LDAP Link.
func resourceGitLabGroupLDAPLinkBuildId(groupId string, ldapProvider string, cn string, filter string) string {
	return fmt.Sprintf("%s:%s:%s:%s", groupId, ldapProvider, cn, filter)
}

// Parses the GitLabGroupLDAPLink ID, which uses a 3-part ID as opposed to the more "normal" 2 part ID.
func resourceGitLabGroupLDAPLinkParseId(id string) (string, string, string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 4 {
		return "", "", "", "", errors.New("unexpected ID format: Group LDAP Link ID had fewer than 4 parts. Expected <GroupID>:<LDAPProvider>:<CN>:<filter>")
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}

func resourceGitlabGroupLdapLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)

	group := d.Get("group").(string)
	cn := d.Get("cn").(string)
	filter := d.Get("filter").(string)

	var groupAccess gitlab.AccessLevelValue
	if v, ok := d.GetOk("group_access"); ok {
		groupAccess = gitlab.AccessLevelValue(api.AccessLevelNameToValue[v.(string)])
	} else if v, ok := d.GetOk("access_level"); ok {
		groupAccess = gitlab.AccessLevelValue(api.AccessLevelNameToValue[v.(string)])
	} else {
		return diag.Errorf("Neither `group_access` nor `access_level` (deprecated) is set")
	}

	ldap_provider := d.Get("ldap_provider").(string)
	force := d.Get("force").(bool)

	options := &gitlab.AddGroupLDAPLinkOptions{
		GroupAccess: &groupAccess,
		Provider:    &ldap_provider,
	}
	if cn != "" {
		options.CN = &cn
	}
	if filter != "" {
		options.Filter = &filter
	}

	if force {
		if err := resourceGitlabGroupLdapLinkDelete(ctx, d, meta); err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] Create GitLab group LdapLink %s", d.Id())
	ldapLink, _, err := client.Groups.AddGroupLDAPLink(group, options, gitlab.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resourceGitLabGroupLDAPLinkBuildId(group, ldapLink.Provider, ldapLink.CN, ldapLink.Filter))
	return resourceGitlabGroupLdapLinkRead(ctx, d, meta)
}

func resourceGitlabGroupLdapLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)
	group, ldapProvider, cn, filter, err := resourceGitLabGroupLDAPLinkParseId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Try to fetch all group links from GitLab
	log.Printf("[DEBUG] Read GitLab group LdapLinks %s", group)
	ldapLinks, _, err := client.Groups.ListGroupLDAPLinks(group, nil, gitlab.WithContext(ctx))
	if err != nil {
		// NOTE: the LDAP list API returns a 404 if there are no LDAP links present.
		if !api.Is404(err) {
			return diag.FromErr(err)
		}
	}

	found := false
	// Check if the LDAP link exists in the returned list of links
	for _, ldapLink := range ldapLinks {
		if ldapProvider == ldapLink.Provider &&
			cn == ldapLink.CN &&
			filter == ldapLink.Filter {

			d.Set("group", group)
			d.Set("cn", ldapLink.CN)
			d.Set("group_access", api.AccessLevelValueToName[ldapLink.GroupAccess])
			d.Set("ldap_provider", ldapLink.Provider)
			d.Set("filter", ldapLink.Filter)
			found = true
			break
		}
	}

	if !found {
		d.SetId("")
		log.Printf("LdapLink %s does not exist, removing from state.", d.Id())
		return nil
	}

	return nil
}

func resourceGitlabGroupLdapLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)
	group, ldap_provider, cn, filter, err := resourceGitLabGroupLDAPLinkParseId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Delete GitLab group LdapLink %s", d.Id())
	options := gitlab.DeleteGroupLDAPLinkWithCNOrFilterOptions{
		Provider: &ldap_provider,
	}
	if cn != "" {
		options.CN = &cn
	}
	if filter != "" {
		options.Filter = &filter
	}

	_, err = client.Groups.DeleteGroupLDAPLinkWithCNOrFilter(group, &options, gitlab.WithContext(ctx))
	if err != nil {

		switch err.(type) { // nolint // TODO: Resolve this golangci-lint issue: S1034: assigning the result of this type assertion to a variable (switch err := err.(type)) could eliminate type assertions in switch cases (gosimple)
		case *gitlab.ErrorResponse:
			// Ignore LDAP links that don't exist
			if strings.Contains(string(err.(*gitlab.ErrorResponse).Message), "Linked LDAP group not found") { // nolint // TODO: Resolve this golangci-lint issue: S1034(related information): could eliminate this type assertion (gosimple)
				log.Printf("[WARNING] %s", err)
			} else {
				return diag.FromErr(err)
			}
		default:
			return diag.FromErr(err)
		}
	}

	return nil
}
