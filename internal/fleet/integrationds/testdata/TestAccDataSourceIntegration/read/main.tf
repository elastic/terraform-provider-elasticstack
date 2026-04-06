provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_fleet_integration" "test" {
  name = "tcp"
}
