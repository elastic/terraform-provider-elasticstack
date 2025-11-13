
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_kibana_spaces" "all_spaces" {}
