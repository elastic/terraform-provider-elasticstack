provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_case_template" "test" {
  owner = "cases"
  name  = "t"
  definition_yaml = format("name: x\nfields: []\n# %s", join("", [for i in range(1000) : "0123456789012345678901234567890"]))
}
