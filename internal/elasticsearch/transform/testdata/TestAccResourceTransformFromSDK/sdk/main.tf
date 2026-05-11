variable "transform_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

# PF-shaped config matching TestAccResourceTransformFromSDK/main.tf (must stay
# in sync for ExpectEmptyPlan after v0→v1 state upgrade).
resource "elasticstack_elasticsearch_transform" "test" {
  name        = var.transform_name
  description = "test transform from sdk"

  source {
    indices = ["source_index_for_transform"]
  }

  destination {
    index = "dest_index_for_transform_from_sdk"
  }

  pivot = jsonencode({
    "group_by" : {
      "customer_id" : {
        "terms" : {
          "field" : "customer_id"
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
      max_age = "30d"
    }
  }

  frequency        = "5m"
  enabled          = false
  defer_validation = true
}
