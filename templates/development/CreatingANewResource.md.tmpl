# Creating a New Resource Using the Terraform Plugin Framework

This tutorial is meant to help new contributors out when creating new resource. It will walk through a 
step-by-step guide of creating a new resource using the 
[Terraform Provider Framework](https://developer.hashicorp.com/terraform/plugin/framework),
since that is how all new resources are added to the GitLab terraform provider, as noted in the
[CONTRIBUTING.md](/CONTRIBUTING.md). This guide will assume that a development environment has already
been set up by following the `Developing The Provider` section of the CONTRIBUTING.md documentation.

<!-- Use "yzhang.markdown-all-in-one" plugin to keep this up to date in vscode -->
- [Creating a New Resource Using the Terraform Plugin Framework](#creating-a-new-resource-using-the-terraform-plugin-framework)
	- [Step 1: Understand the API from GitLab](#step-1-understand-the-api-from-gitlab)
	- [Step 2: Create the Resource struct](#step-2-create-the-resource-struct)
	- [Step 3: Create the Schema for the Resource](#step-3-create-the-schema-for-the-resource)


## Step 1: Understand the API from GitLab

When creating a new resource, the GitLab terraform provider follows the
[Terraform Provider Best Practices](https://developer.hashicorp.com/terraform/plugin/best-practices/hashicorp-provider-design-principles)
whenever possible. This means that a new resource meets a couple of criteria:

1. One resource aligns as closely to one set of CRUD APIs as possible.
2. The attributes of the resource align to the attributes of the underlying APIs.

For this example, the [`resource_gitlab_application`](../internal/provider/resource_gitlab_application.go)
resource will be used as a step-by-step example. This resource aligns to the 
[Applications API](https://docs.gitlab.com/ee/api/applications.html) exposed by GitLab. When creating
a resource, first ensure that the relevant APIs are present in GitLab. If it's not clear whether an
api exists for a resource, create an issue on the GitLab Terraform Provider project and ask!

## Step 2: Create the Resource struct

In the Terraform Plugin framework, each resource is represented by a struct that implements one or more
interfaces. For the sake of keeping this tutorial simple, these interfaces won't be covered in details. However,
creating the resource struct will be the first step in creating a new resource. Each resource is created
within its own `go` file, named `resource_<resource_name>.go`; in this case, `resource_gitlab_application.go`. 

```golang
type gitlabApplicationResource struct {
	client *gitlab.Client // This is required for making calls to GitLab later
}
```

Creating the struct alone isn't enough to ensure the interfaces are met, so for error handling reasons, a 
block at the top of th

## Step 3: Create the Schema for the Resource

The schema for the resource handles multiple responsibilities during `terraform plan` and `terraform apply`:

1. It ensures that the input data is the correct type (`number` vs `string`).
2. It ensures that the input data is properly validated (matches any validation rules).
3. It ensures that the input data has all the necessarily required fields.

As a result, the schema is the natural starting point for creating a resource. The best place to start
for creating a resource is to copy all the required and optional attributes from the GitLab API into the
schema struct. To define the schema for the resource, first, create a struct representing the attributes
that a user can use to configure the resource:

```golang
type gitlabApplicationResourceModel struct {
	Name         types.String `tfsdk:"name"`
	RedirectURL  types.String `tfsdk:"redirect_url"`
	Scopes       types.Set    `tfsdk:"scopes"`
	Confidential types.Bool   `tfsdk:"confidential"`

	Id            types.String `tfsdk:"id"`
	Secret        types.String `tfsdk:"secret"`
	ApplicationId types.String `tfsdk:"application_id"`
}
```
 
There are a couple of things to notice about this struct:

1. The types for each attribute use the `types` package. This is because `types.String` can have a nil value,
whereas a primative `string` cannot.
2. The `tfsdk` tag value maps to the string value in our schema.
3. Each new struct like this must have a unique name. The terraform provider uses the naming convention of
`gitlab<resourceName><resource type, either Resource or Data>Model`. That means an application data source 
would be named `gitlabApplicationDataModel`.

After the schema struct is created, the next step is to create a second struct representing the resource itself. This
struct will then implement all the functions that are required for performing terraform CRUD (Create, Read,
Update, Delete) operations.

```golang
type gitlabApplicationResource struct {
	client *gitlab.Client
}
```

This struct is very simple, and just accepts a client reference. This client will be used to make REST calls to
the GitLab instance configured in the provider.

With the schema struct and the resource struct created, it's time to start implementing the resource functions.

The first function to create is the `Schema` function, which defines a `schema.Schema` struct representing the schema
and all the validations required for the resource. The schema block is very large, so the full block will not be copied here. 
The full schema function can be read 
[in the repository, linked here](https://gitlab.com/gitlab-org/terraform-provider-gitlab/-/blob/main/internal/provider/resource_gitlab_application.go#L63)

```golang
func (r *gitlabApplicationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: fmt.Sprintf(`The ` + "`gitlab_application`" + ` resource allows to manage the lifecycle of applications in gitlab.

~> In order to use a user for a user to create an application, they must have admin priviledges at the instance level.
To create an OIDC application, a scope of "openid".

**Upstream API**: [GitLab REST API docs](https://docs.gitlab.com/ee/api/applications.html)`),

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this Terraform resource. In the format of `<application_id>`.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the application.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			// additional schema resources past this point.
		}
	}
}
```

Similar to the schema struct above, there are a couple things to take note of in the above `Schema` func.

1. The `Schema` func itself is part of the `resource.Resource` interface. Make sure it has the proper inputs!
2. Each `Schema` must have a `MarkdownDescription`. This will appear in the terraform documentation on the provider's site.
3. Each `Schema` must have a `Attributes` map, which contains a minimum of one `schema.Attribute` in its map. This map
is where plan-time validation happens. Within each `schema.Attribute`, several key properties are required:
  - `MarkdownDescription` if the documentation that will appear on the terraform documentation site for that attribute.
  - `Required` denotes whether the attribute is required for the resource. Resources missing required attribute will fail at plan-time.
  - `Computed` denotes whether the resource will compute values for that attribute that may differ from the plan. If `Computed` is 
  set to `true`, then storing a value that's different from the terraform config won't result in a diff being identified unless the 
  value is explicitly set in the config. 
  - `Validators` accepts validator functions that can be used to validate inputs at plan time.
  - `PlanModifiers` accepts modifier functions that can change how the resource identifies plan changes.

For more information on various properties of the schema attributes, feel free to read the 
[Terraform Plugin Framework Schema Documentation](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas).

After the schema function has been written, the `Config` function needs to be written. Don't worry, this one is much easier!

```golang
// Configure adds the provider configured client to the resource.
func (r *gitlabApplicationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*gitlab.Client)
}
```

This function will be nearly identical on every resource. The logic simply sets the client in the resource struct to be the value 
configured in the provider. This ensures that when making calls from the `r.Client` that they're authenticated and configured properly.

Finally, it's time to create the CRUD functions for the resource. The CRUD functions (Create, Read, Update, and Delete) are responsible
for using the `r.Client` to make the changes to the GitLab instance. Terraform will automatically call the correct function based on 
the terraform plan that's generated before the apply:

- If a resource is labelled as `create`, the `Create` function will be called. 
- - If a resource is labelled as `update`, the `Update` function will be called.
- If a resource is labelled as `destroy`, the `Delete` function will be called. 
- The `Read` function is called any time terraform `refresh` is called, either by a `plan`, an `apply`, or an explicit `refresh`.