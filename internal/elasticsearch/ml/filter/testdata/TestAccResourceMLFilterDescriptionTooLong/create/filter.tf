variable "filter_id" {
  description = "The filter ID"
  type        = string
}

locals {
  # Terraform's range() is capped at 1024 values per call; build 4097 chars in chunks.
  desc_chunk = join("", [for i in range(1024) : "a"])
  too_long_description = join("", [
    local.desc_chunk,
    local.desc_chunk,
    local.desc_chunk,
    local.desc_chunk,
    "a",
  ])
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_filter" "test" {
  filter_id   = var.filter_id
  description = local.too_long_description
  items       = ["*.example.com"]
}
