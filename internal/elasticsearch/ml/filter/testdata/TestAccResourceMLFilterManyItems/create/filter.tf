variable "filter_id" {
  description = "The filter ID"
  type        = string
}

locals {
  # Many distinct items to exercise update diff / API payload size without hitting the 10000-item limit.
  many_items = [for i in range(250) : format("item-%04d.example.com", i)]
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_filter" "test" {
  filter_id   = var.filter_id
  description = "Filter with many items"
  items       = local.many_items
}
