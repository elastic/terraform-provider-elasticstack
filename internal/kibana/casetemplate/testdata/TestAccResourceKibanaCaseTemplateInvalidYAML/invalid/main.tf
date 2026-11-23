provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_case_template" "test" {
  owner           = "cases"
  name            = "t"
  definition_yaml = <<-EOT
a: [unterminated
EOT
}
