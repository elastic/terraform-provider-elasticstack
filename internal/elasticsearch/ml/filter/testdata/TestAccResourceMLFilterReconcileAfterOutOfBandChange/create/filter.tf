variable "filter_id" {
  description = "The filter ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_filter" "test" {
  filter_id   = var.filter_id
  description = "Baseline description for drift reconcile"
  items       = ["*.example.com", "trusted.org", "*.safe.net"]
}
