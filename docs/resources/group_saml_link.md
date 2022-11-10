---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "gitlab_group_saml_link Resource - terraform-provider-gitlab"
subcategory: ""
description: |-
  The gitlab_group_saml_link resource allows to manage the lifecycle of an SAML integration with a group.
  Upstream API: GitLab REST API docs https://docs.gitlab.com/ee/api/groups.html#saml-group-links
---

# gitlab_group_saml_link (Resource)

The `gitlab_group_saml_link` resource allows to manage the lifecycle of an SAML integration with a group.

**Upstream API**: [GitLab REST API docs](https://docs.gitlab.com/ee/api/groups.html#saml-group-links)

## Example Usage

```terraform
resource "gitlab_group_saml_link" "test" {
  group           = "12345"
  access_level    = "developer"
  saml_group_name = "samlgroupname1"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `access_level` (String) Access level for members of the SAML group. Valid values are: `guest`, `reporter`, `developer`, `maintainer`, `owner`.
- `group` (String) The ID or path of the group to add the SAML Group Link to.
- `saml_group_name` (String) The name of the SAML group.

### Read-Only

- `id` (String) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
# GitLab group saml links can be imported using an id made up of `group_id:saml_group_name`, e.g.
terraform import gitlab_group_saml_link.test "12345:samlgroupname1"
```