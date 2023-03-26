# Creating a New Resource Using the Terraform Plugin Framework

This tutorial is meant to help new contributors out when creating new resource. It will walk through a 
step-by-step guide of creating a new resource using the 
[Terraform Provider Framework](https://developer.hashicorp.com/terraform/plugin/framework),
since that is how all new resources are added to the GitLab terraform provider, as noted in the
[CONTRIBUTING.md](/CONTRIBUTING.md). This guide will assume that a development environment has already
been set up by following the `Developing The Provider` section of the CONTRIBUTING.md documentation.

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
schema struct. To define the schema for the resource, first, create

```golang


```
