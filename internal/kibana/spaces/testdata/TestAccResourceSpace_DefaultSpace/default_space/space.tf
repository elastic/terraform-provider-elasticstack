provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "default" {
  space_id    = "default"
  name        = "Default"
  description = "This is your default space!"
}
