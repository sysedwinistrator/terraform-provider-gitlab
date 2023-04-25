resource "gitlab_project_label" "fixme" {
  project     = "example"
  name        = "fixme"
  description = "issue with failing tests"
  color       = "#ffcc00"
}

# Scoped label
resource "gitlab_project_label" "devops_create" {
  project     = gitlab_project.example.id
  name        = "devops::create"
  description = "issue for creating infrastructure resources"
  color       = "#ffa500"
}

