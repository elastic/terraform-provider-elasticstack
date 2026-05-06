provider "elasticstack" {
  kibana {}
}

# Example 1: Upload a source map using inline JSON content.
# The sourcemap_json value is write-only and never read back from the API.
resource "elasticstack_apm_source_map" "example_json" {
  service_name    = "my-frontend"
  service_version = "1.0.0"
  bundle_filepath = "/static/js/main.chunk.js"
  sourcemap_json = jsonencode({
    version  = 3
    file     = "main.chunk.js"
    sources  = ["src/index.js"]
    mappings = "AAAA"
  })
}

# Example 2: Upload a source map using base64-encoded binary content,
# scoped to a non-default Kibana space.
# The sourcemap_binary value is write-only and never read back from the API.
resource "elasticstack_apm_source_map" "example_space" {
  service_name     = "my-frontend"
  service_version  = "2.0.0"
  bundle_filepath  = "/static/js/main.chunk.js"
  sourcemap_binary = filebase64("path/to/main.chunk.js.map")
  space_id         = "my-space"
}
