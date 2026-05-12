variable "transform_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_transform" "test" {
  name        = var.transform_name
  description = "sync time without field"

  source {
    indices = ["source_index_for_transform"]
  }

  destination {
    index = "dest_index_for_transform"
  }

  pivot = jsonencode({
    group_by = {
      customer_id = {
        terms = { field = "customer_id" }
      }
    }
  })

  sync {
    time {
      delay = "20s"
    }
  }

  defer_validation = true
}
