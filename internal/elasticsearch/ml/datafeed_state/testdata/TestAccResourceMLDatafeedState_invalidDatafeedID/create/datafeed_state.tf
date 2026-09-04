variable "invalid_datafeed_id" {
  description = "An invalid ML datafeed ID (contains a disallowed character)"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_datafeed_state" "test" {
  datafeed_id = var.invalid_datafeed_id
  state       = "started"
}
