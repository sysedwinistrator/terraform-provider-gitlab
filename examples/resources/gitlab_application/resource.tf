resource "gitlab_application" "oidc" {
  confidential = true
  scopes       = ["openid"]
  name         = "company_oidc"
  redirect_url = "https://mycompany.com"
}