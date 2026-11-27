provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

data "elasticstack_fleet_server_host" "test" {
  host_id = "does-not-exist-${var.suffix}"
}
