# To search for a project by ID, pass in the ID value
data "gitlab_project" "example" {
  id = 30
}

# To search for a project based on a path, use `path_with_namespace` instead
data "gitlab_project" "example" {
  path_with_namespace = "foo/bar/baz"
}
