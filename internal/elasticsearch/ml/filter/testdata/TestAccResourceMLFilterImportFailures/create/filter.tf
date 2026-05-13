variable "filter_id" {
  description = "The filter ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_filter" "test" {
  filter_id   = var.filter_id
  description = "Filter for import failure tests"
  items       = ["*.example.com", "trusted.org"]
}
