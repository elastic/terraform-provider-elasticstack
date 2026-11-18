variable "job_id_a" {
  description = "The ML job ID for datafeed A"
  type        = string
}

variable "job_id_b" {
  description = "The ML job ID for datafeed B"
  type        = string
}

variable "datafeed_id_a" {
  description = "The ML datafeed ID for datafeed A"
  type        = string
}

variable "datafeed_id_b" {
  description = "The ML datafeed ID for datafeed B"
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

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "a" {
  job_id      = var.job_id_a
  description = "Job A for datafeed_id ForceNew test"
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

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "b" {
  job_id      = var.job_id_b
  description = "Job B for datafeed_id ForceNew test"
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

resource "elasticstack_elasticsearch_ml_job_state" "a" {
  job_id = elasticstack_elasticsearch_ml_anomaly_detection_job.a.job_id
  state  = "opened"
}

resource "elasticstack_elasticsearch_ml_job_state" "b" {
  job_id = elasticstack_elasticsearch_ml_anomaly_detection_job.b.job_id
  state  = "opened"
}

resource "elasticstack_elasticsearch_ml_datafeed" "a" {
  datafeed_id = var.datafeed_id_a
  job_id      = elasticstack_elasticsearch_ml_anomaly_detection_job.a.job_id
  indices     = [elasticstack_elasticsearch_index.test.name]
  query = jsonencode({
    match_all = {}
  })
}

resource "elasticstack_elasticsearch_ml_datafeed" "b" {
  datafeed_id = var.datafeed_id_b
  job_id      = elasticstack_elasticsearch_ml_anomaly_detection_job.b.job_id
  indices     = [elasticstack_elasticsearch_index.test.name]
  query = jsonencode({
    match_all = {}
  })
}

resource "elasticstack_elasticsearch_ml_datafeed_state" "test" {
  datafeed_id = elasticstack_elasticsearch_ml_datafeed.b.datafeed_id
  state       = "stopped"

  depends_on = [
    elasticstack_elasticsearch_ml_datafeed.a,
    elasticstack_elasticsearch_ml_datafeed.b,
    elasticstack_elasticsearch_ml_job_state.a,
    elasticstack_elasticsearch_ml_job_state.b,
  ]
}
