terraform {
  required_providers {
    gitlab = {
      source = "gitlabhq/gitlab"
    }
  }
}

provider "gitlab" {
  base_url = "http://localhost:8085"
  token    = "glpat-ACCTEST1234567890123"
}

data "gitlab_metadata" "this" {}

data "gitlab_current_user" "this" {}
