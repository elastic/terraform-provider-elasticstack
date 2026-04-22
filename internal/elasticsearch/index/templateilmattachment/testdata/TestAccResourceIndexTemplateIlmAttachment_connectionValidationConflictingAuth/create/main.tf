provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template_ilm_attachment" "test" {
  index_template = "validation-conflicting-auth"
  lifecycle_name = "validation-conflicting-auth"

  elasticsearch_connection {
    endpoints = ["http://localhost:9200"]
    username  = "elastic"
    password  = "password"
    api_key   = "Zm9vOmJhcg=="
  }
}
