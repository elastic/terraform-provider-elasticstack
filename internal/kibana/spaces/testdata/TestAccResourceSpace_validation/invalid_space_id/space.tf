provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id = "Invalid Space ID!"
  name     = "Invalid Space ID"
}
