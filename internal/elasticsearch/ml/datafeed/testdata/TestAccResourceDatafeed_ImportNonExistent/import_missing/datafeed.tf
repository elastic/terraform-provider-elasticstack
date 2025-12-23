variable "job_id" {
  type = string
}

variable "datafeed_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_datafeed" "test" {
  datafeed_id = var.datafeed_id
  job_id      = var.job_id
  indices     = ["test-index-*"]

  query = jsonencode({
    match_all = {
      boost = 1
    }
  })
}
