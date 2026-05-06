variable "service_name" {
  description = "The APM service name"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_apm_source_map" "test" {
  bundle_filepath = "/static/js/test.min.js"
  service_name    = var.service_name
  service_version = "1.1.0"
  sourcemap = {
    json = "{\"version\":3,\"file\":\"test.min.js\",\"sources\":[\"test.js\"],\"mappings\":\"AAAA\"}"
  }
}
