variable "name" {
  type = string
}

variable "host_id" {
  type = string
}

variable "second_space_id" {
  type = string
}

variable "third_space_id" {
  type = string
}

variable "second_space_name" {
  type = string
}

variable "third_space_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "second" {
  space_id = var.second_space_id
  name     = var.second_space_name
}

resource "elasticstack_kibana_space" "third" {
  space_id = var.third_space_id
  name     = var.third_space_name
}

resource "elasticstack_fleet_server_host" "test_host" {
  name    = var.name
  host_id = var.host_id
  default = false
  hosts = [
    "https://fleet-server:8220"
  ]
  # The plan set must not include the prior operational space ("default").
  # space_ids is a Set, so iteration order is not stable; if update.go used
  # the plan instead of prior state to resolve the operational space, neither
  # element here would be the space where the resource currently exists.
  space_ids = [var.second_space_id, var.third_space_id]

  depends_on = [
    elasticstack_kibana_space.second,
    elasticstack_kibana_space.third,
  ]
}
