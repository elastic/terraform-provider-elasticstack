provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_case_template" "test" {
  owner = "cases"
  name  = "${join("", [for i in range(5) : "0123456789"])}x"
  definition_yaml = "name: x\nfields: []\n"
}
