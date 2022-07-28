provider "elasticstack" {
  elasticsearch {
    api_key   = "base64encodedapikeyhere=="
    endpoints = ["http://localhost:9200"]
  }
}
