provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_advanced_settings" "test" {
  global = true
  settings = {
    "xpackCustomBranding:pageTitle" = jsonencode("Terraform")
  }
}
