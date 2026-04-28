provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id    = var.space_id
  name        = var.space_name
  description = "Test space for fleet server host space import test"
}

resource "elasticstack_fleet_server_host" "test_host" {
  name    = var.name
  host_id = "fleet-server-host-id"
  default = false
  hosts = [
    "https://fleet-server:8220"
  ]
  space_ids = [var.space_id]

  depends_on = [elasticstack_kibana_space.test_space]
}
