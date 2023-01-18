package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/xanzy/go-gitlab"
	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/api"
	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &gitlabProjectProtectedEnvironmentResource{}
var _ resource.ResourceWithConfigure = &gitlabProjectProtectedEnvironmentResource{}
var _ resource.ResourceWithImportState = &gitlabProjectProtectedEnvironmentResource{}

func init() {
	registerResource(NewGitLabProjectProtectedEnvironmentResource)
}

// NewGitLabProjectProtectedEnvironmentResource is a helper function to simplify the provider implementation.
func NewGitLabProjectProtectedEnvironmentResource() resource.Resource {
	return &gitlabProjectProtectedEnvironmentResource{}
}

// gitlabProjectProtectedEnvironmentResource defines the resource implementation.
type gitlabProjectProtectedEnvironmentResource struct {
	client *gitlab.Client
}

// gitlabProjectProtectedEnvironmentResourceModel describes the resource data model.
type gitlabProjectProtectedEnvironmentResourceModel struct {
	Id                    types.String                                              `tfsdk:"id"`
	Project               types.String                                              `tfsdk:"project"`
	Environment           types.String                                              `tfsdk:"environment"`
	RequiredApprovalCount types.Int64                                               `tfsdk:"required_approval_count"`
	DeployAccessLevels    []gitlabProjectProtectedEnvironmentDeployAccessLevelModel `tfsdk:"deploy_access_levels"`
}

type gitlabProjectProtectedEnvironmentDeployAccessLevelModel struct {
	AccessLevel            types.String `tfsdk:"access_level"`
	AccessLevelDescription types.String `tfsdk:"access_level_description"`
	UserId                 types.Int64  `tfsdk:"user_id"`
	GroupId                types.Int64  `tfsdk:"group_id"`
}

func (r *gitlabProjectProtectedEnvironmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_protected_environment"
}

func (r *gitlabProjectProtectedEnvironmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `The ` + "`gitlab_project_protected_environment`" + ` resource allows to manage the lifecycle of a protected environment in a project.

~> In order to use a user or group in the ` + "`deploy_access_levels`" + ` configuration,
   you need to make sure that users have access to the project and groups must have this project shared.
   You may use the ` + "`gitlab_project_membership`" + ` and ` + "`gitlab_project_shared_group`" + ` resources to achieve this.
   Unfortunately, the GitLab API does not complain about users and groups without access to the project and just ignores those.
   In case this happens you will get perpetual state diffs.

**Upstream API**: [GitLab REST API docs](https://docs.gitlab.com/ee/api/protected_environments.html)`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of this Terraform resource. In the format of `<project>:<environment-name>`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"project": schema.StringAttribute{
				MarkdownDescription: "The ID or full path of the project which the protected environment is created against.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "The name of the environment.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"required_approval_count": schema.Int64Attribute{
				MarkdownDescription: "The number of approvals required to deploy to this environment.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"deploy_access_levels": schema.SetNestedBlock{
				MarkdownDescription: "Array of access levels allowed to deploy, with each described by a hash.",
				Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
				PlanModifiers:       []planmodifier.Set{setplanmodifier.RequiresReplace(), setplanmodifier.UseStateForUnknown()},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"access_level": schema.StringAttribute{
							MarkdownDescription: fmt.Sprintf("Levels of access required to deploy to this protected environment. Valid values are %s.", utils.RenderValueListForDocs(api.ValidProtectedEnvironmentDeploymentLevelNames)),
							Optional:            true,
							PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
							Validators: []validator.String{
								stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("user_id"), path.MatchRelative().AtParent().AtName("group_id")),
								stringvalidator.OneOfCaseInsensitive(api.ValidProtectedEnvironmentDeploymentLevelNames...),
							},
						},
						"access_level_description": schema.StringAttribute{
							MarkdownDescription: "Readable description of level of access.",
							Computed:            true,
							PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						},
						"user_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the user allowed to deploy to this protected environment. The user must be a member of the project.",
							Optional:            true,
							PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
							Validators:          []validator.Int64{int64validator.AtLeast(1)},
						},
						"group_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the group allowed to deploy to this protected environment. The project must be shared with the group.",
							Optional:            true,
							PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
							Validators:          []validator.Int64{int64validator.AtLeast(1)},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *gitlabProjectProtectedEnvironmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*gitlab.Client)
}

// Create creates a new upstream resources and adds it into the Terraform state.
func (r *gitlabProjectProtectedEnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *gitlabProjectProtectedEnvironmentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// local copies of plan arguments
	projectID := data.Project.ValueString()
	environmentName := data.Environment.ValueString()

	// configure GitLab API call
	options := &gitlab.ProtectRepositoryEnvironmentsOptions{
		Name: gitlab.String(environmentName),
	}

	if !data.RequiredApprovalCount.IsNull() {
		options.RequiredApprovalCount = gitlab.Int(int(data.RequiredApprovalCount.ValueInt64()))
	}

	// deploy access levels
	deployAccessLevelsOption := make([]*gitlab.EnvironmentAccessOptions, len(data.DeployAccessLevels))
	for i, v := range data.DeployAccessLevels {
		deployAccessLevelOptions := &gitlab.EnvironmentAccessOptions{}

		if !v.AccessLevel.IsNull() && v.AccessLevel.ValueString() != "" {
			deployAccessLevelOptions.AccessLevel = gitlab.AccessLevel(api.AccessLevelNameToValue[v.AccessLevel.ValueString()])
		}
		if !v.UserId.IsNull() && v.UserId.ValueInt64() != 0 {
			deployAccessLevelOptions.UserID = gitlab.Int(int(v.UserId.ValueInt64()))
		}
		if !v.GroupId.IsNull() && v.GroupId.ValueInt64() != 0 {
			deployAccessLevelOptions.GroupID = gitlab.Int(int(v.GroupId.ValueInt64()))
		}
		deployAccessLevelsOption[i] = deployAccessLevelOptions
	}
	options.DeployAccessLevels = &deployAccessLevelsOption

	// Protect environment
	protectedEnvironment, _, err := r.client.ProtectedEnvironments.ProtectRepositoryEnvironments(projectID, options, gitlab.WithContext(ctx))
	if err != nil {
		if api.Is404(err) {
			resp.Diagnostics.AddError(
				"GitLab Feature not available",
				fmt.Sprintf("The protected environment feature is not available on this project. Make sure it's part of an enterprise plan. Error: %s", err.Error()),
			)
			return
		}
		resp.Diagnostics.AddError("GitLab API error occurred", fmt.Sprintf("Unable to protect environment: %s", err.Error()))
		return
	}

	// Create resource ID and persist in state model
	data.Id = types.StringValue(utils.BuildTwoPartID(&projectID, &protectedEnvironment.Name))

	// persist API response in state model
	r.protectedEnvironmentToStateModel(projectID, protectedEnvironment, data)

	// Log the creation of the resource
	tflog.Debug(ctx, "created a protected environment", map[string]interface{}{
		"project": projectID, "environment": environmentName,
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *gitlabProjectProtectedEnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *gitlabProjectProtectedEnvironmentResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// read all information for refresh from resource id
	projectID, environmentName, err := utils.ParseTwoPartID(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid resource ID format",
			fmt.Sprintf("The resource ID '%s' has an invalid format. It should be '<project>:<environment-name>'. Error: %s", data.Id.ValueString(), err.Error()),
		)
		return
	}

	// Read environment protection
	protectedEnvironment, _, err := r.client.ProtectedEnvironments.GetProtectedEnvironment(projectID, environmentName, gitlab.WithContext(ctx))
	if err != nil {
		if api.Is404(err) {
			tflog.Debug(ctx, "protected environment does not exist, removing from state", map[string]interface{}{
				"project": projectID, "environment": environmentName,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("GitLab API error occured", fmt.Sprintf("Unable to read protected environment details: %s", err.Error()))
		return
	}

	// persist API response in state model
	r.protectedEnvironmentToStateModel(projectID, protectedEnvironment, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Updates updates the resource in-place.
func (r *gitlabProjectProtectedEnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Provider Error, report upstream", "Somehow the resource was requested to perform an in-place upgrade which is not possible.")
}

// Deletes removes the resource.
func (r *gitlabProjectProtectedEnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *gitlabProjectProtectedEnvironmentResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// read all information for refresh from resource id
	projectID, environmentName, err := utils.ParseTwoPartID(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid resource ID format",
			fmt.Sprintf("The resource ID '%s' has an invalid format. It should be '<project>:<environment-name>'. Error: %s", data.Id.ValueString(), err.Error()),
		)
		return
	}

	if _, err = r.client.ProtectedEnvironments.UnprotectEnvironment(projectID, environmentName, gitlab.WithContext(ctx)); err != nil {
		resp.Diagnostics.AddError(
			"GitLab API Error occurred",
			fmt.Sprintf("Unable to delete protected environment: %s", err.Error()),
		)
	}
}

// ImportState imports the resource into the Terraform state.
func (r *gitlabProjectProtectedEnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *gitlabProjectProtectedEnvironmentResource) protectedEnvironmentToStateModel(projectID string, protectedEnvironment *gitlab.ProtectedEnvironment, data *gitlabProjectProtectedEnvironmentResourceModel) {
	data.Project = types.StringValue(projectID)
	data.Environment = types.StringValue(protectedEnvironment.Name)
	data.RequiredApprovalCount = types.Int64Value(int64(protectedEnvironment.RequiredApprovalCount))

	deployAccessLevelsData := make([]gitlabProjectProtectedEnvironmentDeployAccessLevelModel, len(protectedEnvironment.DeployAccessLevels))
	for i, v := range protectedEnvironment.DeployAccessLevels {
		deployAccessLevelData := gitlabProjectProtectedEnvironmentDeployAccessLevelModel{
			AccessLevelDescription: types.StringValue(v.AccessLevelDescription),
		}
		if v.AccessLevel != 0 {
			deployAccessLevelData.AccessLevel = types.StringValue(api.AccessLevelValueToName[v.AccessLevel])
		}
		if v.UserID != 0 {
			deployAccessLevelData.UserId = types.Int64Value(int64(v.UserID))
		}
		if v.GroupID != 0 {
			deployAccessLevelData.GroupId = types.Int64Value(int64(v.GroupID))
		}

		deployAccessLevelsData[i] = deployAccessLevelData
	}
	data.DeployAccessLevels = deployAccessLevelsData
}
