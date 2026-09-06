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
  service_version = "1.0.0"
  sourcemap = {
    binary = "eyJ2ZXJzaW9uIjozLCJmaWxlIjoidGVzdC5taW4uanMiLCJzb3VyY2VzIjpbInRlc3QuanMiXSwibWFwcGluZ3MiOiJBQUFBIn0="
  }
}
