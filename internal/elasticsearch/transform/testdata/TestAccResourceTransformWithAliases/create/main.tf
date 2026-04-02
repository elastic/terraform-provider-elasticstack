variable "transform_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_transform" "test_aliases" {
  name        = var.transform_name
  description = "test transform with aliases"

  source {
    indices = ["source_index"]
  }

  destination {
    index = "dest_index_for_transform"

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
          "field" : "customer_id"
        }
      }
    },
    "aggregations" : {
      "total_sales" : {
        "sum" : {
          "field" : "sales"
        }
      }
    }
  })

  defer_validation = true
}
