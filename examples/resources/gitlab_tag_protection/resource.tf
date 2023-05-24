resource "gitlab_tag_protection" "TagProtect" {
  project             = "12345"
  tag                 = "TagProtected"
  create_access_level = "developer"
  allowed_to_create {
    user_id = 42
  }
  allowed_to_create {
    group_id = 43
  }
}
