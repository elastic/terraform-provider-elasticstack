provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "orders" {
  name           = "orders"
  index_patterns = ["orders-*"]
  priority       = 500

  template {
    settings = jsonencode({
      number_of_shards   = 1
      number_of_replicas = 1
    })
  }
}

resource "elasticstack_elasticsearch_rollover_index_series" "orders" {
  prefix         = "orders"
  rollover_alias = "orders-write"

  # Defaults to true. Destroy deletes every current alias member once this is false.
  deletion_protection = false

  depends_on = [elasticstack_elasticsearch_index_template.orders]
}
