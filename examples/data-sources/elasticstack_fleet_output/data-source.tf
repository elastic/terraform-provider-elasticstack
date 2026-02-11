provider "elasticstack" {
  kibana {}
  fleet {}
}

variable "output_id" {
  type = string
}

data "elasticstack_fleet_output" "example" {
  output_id = var.output_id
}
