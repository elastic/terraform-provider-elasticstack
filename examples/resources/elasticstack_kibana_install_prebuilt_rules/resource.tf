provider "elasticstack" {
  kibana {}
}


resource "elasticstack_kibana_install_prebuilt_rules" "example" {
  space_id = "default"
}
