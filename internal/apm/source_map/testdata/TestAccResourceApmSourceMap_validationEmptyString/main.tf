provider "elasticstack" {
  kibana {}
}

resource "elasticstack_apm_source_map" "test" {
  bundle_filepath = "/static/js/test.min.js"
  service_name    = "my-service"
  service_version = "1.0.0"
  sourcemap = {
    json = ""
  }
}
