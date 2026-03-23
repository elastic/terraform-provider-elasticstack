## The following resources setup a realtime ML datafeed.
resource "elasticstack_elasticsearch_index" "ml_datafeed_index" {
  name = "ml-datafeed-data"
  mappings = jsonencode({
    properties = {
      "@timestamp" = {
        type = "date"
      }
      value = {
        type = "double"
      }
      user = {
        type = "keyword"
      }
    }
  })
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "example" {
  job_id      = "example-anomaly-job"
  description = "Example anomaly detection job"

  analysis_config {
    bucket_span = "15m"
    detectors {
      function      = "mean"
      field_name    = "value"
      by_field_name = "user"
    }
  }

  data_description {
    time_field = "@timestamp"
  }
}

resource "elasticstack_elasticsearch_ml_datafeed" "example" {
  datafeed_id = "example-datafeed"
  job_id      = elasticstack_elasticsearch_ml_anomaly_detection_job.example.job_id
  indices     = [elasticstack_elasticsearch_index.ml_datafeed_index.name]

  query = jsonencode({
    bool = {
      must = [
        {
          range = {
            "@timestamp" = {
              gte = "now-7d"
            }
          }
        }
      ]
    }
  })
}

resource "elasticstack_elasticsearch_ml_datafeed_state" "example" {
  datafeed_id = elasticstack_elasticsearch_ml_datafeed.example.datafeed_id
  state       = "started"
  force       = false
}

## A non-realtime datafeed will automatically stop once all data has been processed.
## It's recommended to ignore changes to the `state` attribute via the resource lifecycle for such datafeeds.

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "non-realtime" {
  job_id      = "non-realtime-anomaly-job"
  description = "Test job for datafeed state testing with time range"
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

resource "elasticstack_elasticsearch_ml_job_state" "non-realtime" {
  job_id = elasticstack_elasticsearch_ml_anomaly_detection_job.non-realtime.job_id
  state  = "opened"

  lifecycle {
    ignore_changes = ["state"]
  }
}

resource "elasticstack_elasticsearch_ml_datafeed" "non-realtime" {
  datafeed_id = "non-realtime-datafeed"
  job_id      = elasticstack_elasticsearch_ml_anomaly_detection_job.non-realtime.job_id
  indices     = [elasticstack_elasticsearch_index.ml_datafeed_index.name]
  query = jsonencode({
    match_all = {}
  })
}

resource "elasticstack_elasticsearch_ml_datafeed_state" "non-realtime" {
  datafeed_id      = elasticstack_elasticsearch_ml_datafeed.non-realtime.datafeed_id
  state            = "started"
  start            = "2024-01-01T00:00:00Z"
  end              = "2024-01-02T00:00:00Z"
  datafeed_timeout = "60s"

  lifecycle {
    ignore_changes = ["state"]
  }
}