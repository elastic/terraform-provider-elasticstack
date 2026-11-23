provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_case_field_definition" "test" {
  owner      = "cases"
  definition = "name: [unclosed"
}
