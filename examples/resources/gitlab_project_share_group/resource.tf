resource "gitlab_project_share_group" "test" {
  project      = "12345"
  group_id     = 1337
  group_access = "guest"
}
