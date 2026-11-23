variable "filter_id" {
  description = "The filter ID"
  type        = string
}

locals {
  # More than the SizeAtMost(10000) validator allows. Built via setproduct of two small
  # ranges (rather than a single range(10001)) because Terraform's range() function
  # refuses to generate more than 1024 values in one call.
  too_many_items = [for pair in setproduct(range(101), range(100)) : format("item-%03d-%03d", pair[0], pair[1])]
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_filter" "test" {
  filter_id   = var.filter_id
  description = "Filter with too many items"
  items       = local.too_many_items
}
