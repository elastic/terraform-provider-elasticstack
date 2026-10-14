provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_case_template" "test" {
  owner           = "cases"
  name            = "t"
  description     = join("", [for i in range(101) : "0123456789"])
  definition_yaml = "name: x\nfields: []\n"
}
