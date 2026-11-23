provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_case_field_definition" "test" {
  owner      = "not-a-solution"
  definition = "name: f\ntype: keyword"
}
