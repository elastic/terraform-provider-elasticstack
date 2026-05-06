provider "elasticstack" {
  kibana {}
}

# Example 1: Upload a source map using inline JSON content.
# sourcemap.json is write-only and never read back from the API.
resource "elasticstack_apm_source_map" "example_json" {
  service_name    = "my-frontend"
  service_version = "1.0.0"
  bundle_filepath = "/static/js/main.chunk.js"
  sourcemap = {
    json = jsonencode({
      version  = 3
      file     = "main.chunk.js"
      sources  = ["src/index.js"]
      mappings = "AAAA"
    })
  }
}

# Example 2: Upload a source map using base64-encoded binary content,
# scoped to a non-default Kibana space.
# sourcemap.binary is write-only and never read back from the API.
resource "elasticstack_apm_source_map" "example_space" {
  service_name    = "my-frontend"
  service_version = "2.0.0"
  bundle_filepath = "/static/js/main.chunk.js"
  sourcemap = {
    # base64-encoded source map content ({"version":3,"file":"main.chunk.js","sources":["src/index.js"],"mappings":"AAAA"})
    binary = "eyJ2ZXJzaW9uIjozLCJmaWxlIjoibWFpbi5jaHVuay5qcyIsInNvdXJjZXMiOlsic3JjL2luZGV4LmpzIl0sIm1hcHBpbmdzIjoiQUFBQSJ9"
  }
  space_id = "my-space"
}

# Example 3: Upload a source map from a local file.
# sourcemap.file.checksum is computed automatically from the file contents.
resource "elasticstack_apm_source_map" "example_file" {
  service_name    = "my-frontend"
  service_version = "3.0.0"
  bundle_filepath = "/static/js/main.chunk.js"
  sourcemap = {
    file = {
      path = "/path/to/main.chunk.js.map"
    }
  }
}
