variable "transform_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_transform" "test_pivot" {
  name        = var.transform_name
  description = "test description"

  source {
    indices = ["source_index_for_transform"]
  }

  destination {
    index = "dest_index_for_transform"
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

  max_page_search_size = 2000
  frequency            = "5m"
  enabled              = false

  defer_validation = true
  timeout          = "1m"
}
