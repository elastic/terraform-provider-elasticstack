variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_server_host" "fleet_host" {
  name    = var.name
  default = false
  hosts   = ["https://fleet-server-issue-864-a.example:8220"]
}
