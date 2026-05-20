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

variable "planned_start" {
  description = "The explicit start timestamp for the datafeed state resource"
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
      "@timestamp" = { type = "date" }
      value        = { type = "double" }
    }
  })
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id      = var.job_id
  description = "Reproducer for issue 2353"
  analysis_config = {
    bucket_span = "15m"
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
  query       = jsonencode({ match_all = {} })
}

# start stays at var.planned_start (e.g. "2022-01-01T00:07:30Z"); effective_search_start
# reflects SearchInterval.StartMs from the first indexed document (e.g. "2022-01-01T00:10:00Z").
# See issue #2353.
resource "elasticstack_elasticsearch_ml_datafeed_state" "test" {
  datafeed_id = elasticstack_elasticsearch_ml_datafeed.test.datafeed_id
  state       = "started"
  start       = var.planned_start

  depends_on = [
    elasticstack_elasticsearch_ml_datafeed.test,
    elasticstack_elasticsearch_ml_job_state.test,
  ]
}
