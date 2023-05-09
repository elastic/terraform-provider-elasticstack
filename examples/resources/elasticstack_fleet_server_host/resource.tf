provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_server_host" "test_host" {
  name    = "Test Host"
  default = false
  hosts = [
    "https://fleet-server:8220"
  ]
}
