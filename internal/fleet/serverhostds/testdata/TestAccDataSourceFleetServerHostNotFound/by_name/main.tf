provider "elasticstack" {
  kibana {}
}

variable "suffix" {
  type = string
}

data "elasticstack_fleet_server_host" "test" {
  name = "does-not-exist-${var.suffix}"
}
