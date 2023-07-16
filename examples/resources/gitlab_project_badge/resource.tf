resource "gitlab_project" "foo" {
  name = "foo-project"
}

resource "gitlab_project_badge" "example" {
  project   = gitlab_project.foo.id
  link_url  = "https://example.com/badge-123"
  image_url = "https://example.com/badge-123.svg"
  name      = "badge-123"
}

# Pipeline status badges with placeholders will be enabled
resource "gitlab_project_badge" "gitlab_pipeline" {
  project   = gitlab_project.foo.id
  link_url  = "https://gitlab.example.com/%%{project_path}/-/pipelines?ref=%%{default_branch}"
  image_url = "https://gitlab.example.com/%%{project_path}/badges/%%{default_branch}/pipeline.svg"
  name      = "badge-pipeline"
}

# Test coverage report badges with placeholders will be enabled
resource "gitlab_project_badge" "gitlab_coverage" {
  project   = gitlab_project.foo.id
  link_url  = "https://gitlab.example.com/%%{project_path}/-/jobs"
  image_url = "https://gitlab.example.com/%%{project_path}/badges/%%{default_branch}/coverage.svg"
  name      = "badge-coverage"
}

# Latest release badges with placeholders will be enabled
resource "gitlab_project_badge" "gitlab_release" {
  project   = gitlab_project.foo.id
  link_url  = "https://gitlab.example.com/%%{project_path}/-/releases"
  image_url = "https://gitlab.example.com/%%{project_path}/-/badges/release.svg"
  name      = "badge-release"
}
