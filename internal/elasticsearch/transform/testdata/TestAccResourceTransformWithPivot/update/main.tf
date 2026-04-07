variable "transform_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test_source_index_1" {
  name = "source_index_for_transform"

  alias = [{
    name = "test_alias_1"
  }]

  mappings = jsonencode({
    properties = {
      field1 = { type = "text" }
    }
  })

  deletion_protection    = false
  wait_for_active_shards = "all"
  master_timeout         = "1m"
  timeout                = "1m"
}

resource "elasticstack_elasticsearch_index" "test_source_index_2" {
  name = "additional_index"

  alias = [{
    name = "test_alias_2"
  }]

  mappings = jsonencode({
    properties = {
      field1 = { type = "text" }
    }
  })

  deletion_protection    = false
  wait_for_active_shards = "all"
  master_timeout         = "1m"
  timeout                = "1m"
}

resource "elasticstack_elasticsearch_transform" "test_pivot" {
  name        = var.transform_name
  description = "yet another test description"

  source {
    indices = [
      elasticstack_elasticsearch_index.test_source_index_1.name,
      elasticstack_elasticsearch_index.test_source_index_2.name
    ]
  }

  destination {
    index = "dest_index_for_transform_v2"
  }

  pivot = jsonencode({
    "group_by" : {
      "customer_id" : {
        "terms" : {
          "field" : "customer_id",
          "missing_bucket" : true
        }
      }
    },
    "aggregations" : {
      "max_price" : {
        "max" : {
          "field" : "taxful_total_price"
        }
      }
    }
  })

  sync {
    time {
      field = "order_date"
      delay = "20s"
    }
  }

  retention_policy {
    time {
      field   = "order_date"
      max_age = "7d"
    }
  }

  max_page_search_size = 2000
  frequency            = "10m"
  enabled              = true

  defer_validation = true
  timeout          = "1m"
}
