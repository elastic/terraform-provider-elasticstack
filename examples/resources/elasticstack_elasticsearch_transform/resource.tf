resource "elasticstack_elasticsearch_transform" "transform_with_pivot" {
  name        = "transform-pivot"
  description = "A meaningful description"

  source {
    indices = ["names_or_patterns_for_input_index"]
  }

  destination {
    index = "destination_index_for_transform"

    aliases {
      alias            = "test_alias_1"
      move_on_creation = true
    }

    aliases {
      alias            = "test_alias_2"
      move_on_creation = false
    }
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

  frequency = "5m"

  retention_policy {
    time {
      field   = "order_date"
      max_age = "30d"
    }
  }

  sync {
    time {
      field = "order_date"
      delay = "10s"
    }
  }

  max_page_search_size = 2000

  enabled          = false
  defer_validation = false
}
