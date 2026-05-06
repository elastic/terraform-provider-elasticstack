variable "service_name" {
  description = "The APM service name"
  type        = string
}

variable "service_version" {
  description = "The APM service version"
  type        = string
}

variable "space_id" {
  description = "The Kibana space ID"
  type        = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "import_space" {
  space_id = var.space_id
  name     = var.space_id
}

resource "elasticstack_apm_source_map" "test" {
  bundle_filepath = "/static/js/test.min.js"
  service_name    = var.service_name
  service_version = var.service_version
  sourcemap = {
    json = "{\"version\":3,\"file\":\"test.min.js\",\"sources\":[\"test.js\"],\"mappings\":\"AAAA\"}"
  }
  space_id = elasticstack_kibana_space.import_space.space_id
}
