variable "filter_id" {
  description = "The filter ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_filter" "test" {
  filter_id   = var.filter_id
  description = "Created by admin for read-only update denial test"
  items       = ["*.example.com"]
}
