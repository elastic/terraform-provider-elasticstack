variable "transform_name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_transform" "test_latest" {
  name        = var.transform_name
  description = "test description (latest)"

  source {
    indices = ["source_index_for_transform"]
  }

  destination {
    index = "dest_index_for_transform"
  }

  latest = jsonencode({
    "unique_key" : ["customer_id"],
    "sort" : "order_date"
  })
  frequency = "2m"
  enabled   = false

  defer_validation = true
  timeout          = "1m"
}
