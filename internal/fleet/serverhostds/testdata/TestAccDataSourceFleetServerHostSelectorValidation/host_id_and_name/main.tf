provider "elasticstack" {
  kibana {}
}

data "elasticstack_fleet_server_host" "test" {
  host_id = "some-host"
  name    = "Some Host"
}
