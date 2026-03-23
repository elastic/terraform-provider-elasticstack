provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_fleet_output" "test" {}