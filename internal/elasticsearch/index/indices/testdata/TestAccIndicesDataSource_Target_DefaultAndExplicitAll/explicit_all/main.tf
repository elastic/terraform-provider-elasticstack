variable "fixture_index" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "fixture" {
  name                = var.fixture_index
  number_of_shards    = 1
  number_of_replicas  = 0
  deletion_protection = false
}

data "elasticstack_elasticsearch_indices" "all_explicit" {
  target = "_all"

  depends_on = [elasticstack_elasticsearch_index.fixture]
}
