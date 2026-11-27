provider "elasticstack" {
  kibana {}
}

data "elasticstack_fleet_server_host" "test" {
  name = ""
}
