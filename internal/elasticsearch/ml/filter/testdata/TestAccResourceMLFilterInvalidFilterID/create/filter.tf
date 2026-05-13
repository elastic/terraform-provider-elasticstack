variable "filter_id" {
  description = "The filter ID (intentionally invalid for this test)"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_filter" "test" {
  filter_id   = var.filter_id
  description = "Invalid filter_id format"
  items       = ["*.example.com"]
}
