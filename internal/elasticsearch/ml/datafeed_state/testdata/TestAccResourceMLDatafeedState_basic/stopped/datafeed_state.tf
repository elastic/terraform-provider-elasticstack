variable "job_id" {
  description = "The ML job ID"
  type        = string
}

variable "datafeed_id" {
  description = "The ML datafeed ID"
  type        = string
}

variable "index_name" {
  description = "The index name"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "test" {
  name                = var.index_name
  deletion_protection = false
  mappings = jsonencode({
    properties = {
      "@timestamp" = {
        type = "date"
      }
      value = {
        type = "double"
      }
    }
  })
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id      = var.job_id
  description = "Test job for datafeed state testing"
  analysis_config = {
    bucket_span = "1h"
    detectors = [{
      function             = "count"
      detector_description = "count"
    }]
  }
  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
  analysis_limits = {
    model_memory_limit = "10mb"
  }
}

resource "elasticstack_elasticsearch_ml_job_state" "test" {
  job_id = elasticstack_elasticsearch_ml_anomaly_detection_job.test.job_id
  state  = "opened"
}

resource "elasticstack_elasticsearch_ml_datafeed" "test" {
  datafeed_id = var.datafeed_id
  job_id      = elasticstack_elasticsearch_ml_anomaly_detection_job.test.job_id
  indices     = [elasticstack_elasticsearch_index.test.name]
  query = jsonencode({
    match_all = {}
  })
}

resource "elasticstack_elasticsearch_ml_datafeed_state" "test" {
  datafeed_id = elasticstack_elasticsearch_ml_datafeed.test.datafeed_id
  state       = "stopped"

  depends_on = [
    elasticstack_elasticsearch_ml_datafeed.test,
    elasticstack_elasticsearch_ml_job_state.test
  ]
}