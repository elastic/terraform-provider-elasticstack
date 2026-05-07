variable "service_name" {
  description = "The APM service name"
  type        = string
}

variable "service_version" {
  description = "The APM service version"
  type        = string
}

variable "file_path" {
  description = "Path to the source map file on the local filesystem"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_apm_source_map" "test" {
  bundle_filepath = "/static/js/test.min.js"
  service_name    = var.service_name
  service_version = var.service_version
  sourcemap = {
    file = {
      path = var.file_path
    }
  }
}
