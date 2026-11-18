variable "datafeed_id" {
  description = "The ML datafeed ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_datafeed_state" "test" {
  datafeed_id = var.datafeed_id
  state       = "starting"
}
