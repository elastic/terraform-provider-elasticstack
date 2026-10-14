provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_case_template" "test" {
  owner           = "cases"
  name            = ""
  definition_yaml = "name: x\nfields: []\n"
}
