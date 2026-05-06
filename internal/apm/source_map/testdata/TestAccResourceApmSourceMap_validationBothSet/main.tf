provider "elasticstack" {
  kibana {}
}

resource "elasticstack_apm_source_map" "test" {
  bundle_filepath = "/static/js/test.min.js"
  service_name    = "my-service"
  service_version = "1.0.0"
  sourcemap = {
    json   = "{\"version\":3,\"file\":\"test.min.js\",\"sources\":[\"test.js\"],\"mappings\":\"AAAA\"}"
    binary = "eyJ2ZXJzaW9uIjozLCJmaWxlIjoidGVzdC5taW4uanMiLCJzb3VyY2VzIjpbInRlc3QuanMiXSwibWFwcGluZ3MiOiJBQUFBIn0="
  }
}
