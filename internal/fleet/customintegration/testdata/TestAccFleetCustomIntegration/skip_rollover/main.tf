variable "package_path" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_custom_integration" "test" {
  package_path              = var.package_path
  skip_data_stream_rollover = true
}
