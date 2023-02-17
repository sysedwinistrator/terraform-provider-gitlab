# Example using auto_ssl_enabled, which uses lets encrypt to generate a certificate
resource "gitlab_pages_domain" "this" {
  project = 123
  domain  = "example.com"

  auto_ssl_enabled = true
}

# Example using a manually generated certificate and key
resource "gitlab_pages_domain" "this" {
  project = 123
  domain  = "example.com"

  key         = file("${path.module}/key.pem")
  certificate = file("${path.module}/cert.pem")
}