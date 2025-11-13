provider "elasticstack" {
  elasticsearch {}
}

# Basic ML Datafeed
resource "elasticstack_elasticsearch_ml_datafeed" "basic" {
  datafeed_id = "my-basic-datafeed"
  job_id      = elasticstack_elasticsearch_ml_anomaly_detection_job.example.job_id
  indices     = ["log-data-*"]

  query = jsonencode({
    match_all = {}
  })
}

# Comprehensive ML Datafeed with all options
resource "elasticstack_elasticsearch_ml_datafeed" "comprehensive" {
  datafeed_id = "my-comprehensive-datafeed"
  job_id      = elasticstack_elasticsearch_ml_anomaly_detection_job.example.job_id
  indices     = ["app-logs-*", "system-logs-*"]

  query = jsonencode({
    bool = {
      must = [
        {
          range = {
            "@timestamp" = {
              gte = "now-1h"
            }
          }
        },
        {
          term = {
            "status" = "error"
          }
        }
      ]
    }
  })

  scroll_size        = 1000
  frequency          = "30s"
  query_delay        = "60s"
  max_empty_searches = 10

  chunking_config = {
    mode      = "manual"
    time_span = "30m"
  }

  delayed_data_check_config = {
    enabled      = true
    check_window = "2h"
  }

  indices_options = {
    ignore_unavailable = true
    allow_no_indices   = false
    expand_wildcards   = ["open", "closed"]
  }

  runtime_mappings = jsonencode({
    "hour_of_day" = {
      "type" = "long"
      "script" = {
        "source" = "emit(doc['@timestamp'].value.getHour())"
      }
    }
  })

  script_fields = jsonencode({
    "my_script_field" = {
      "script" = {
        "source" = "_score * doc['my_field'].value"
      }
    }
  })
}

# Required ML Job for the datafeed
resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "example" {
  job_id      = "example-anomaly-job"
  description = "Example anomaly detection job"

  analysis_config = {
    bucket_span = "15m"
    detectors = [{
      function = "count"
    }]
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
}