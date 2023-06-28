resource "gitlab_compliance_framework" "sample" {
  namespace_path                   = "top-level-group"
  name                             = "HIPAA"
  description                      = "A HIPAA Compliance Framework"
  color                            = "#87BEEF"
  default                          = false
  pipeline_configuration_full_path = ".hipaa.yml@top-level-group/compliance-frameworks"
}
