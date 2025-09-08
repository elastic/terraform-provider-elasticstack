terraform {
  required_providers {
    elasticstack = {
      source  = "elastic/elasticstack"
      version = "~> 0.11"
    }
  }
}

provider "elasticstack" {
  elasticsearch {}
}

# Basic anomaly detection job
resource "elasticstack_elasticsearch_ml_anomaly_detector" "example" {
  job_id      = "example-anomaly-detector"
  description = "Example anomaly detection job for monitoring web traffic"
  groups      = ["web", "monitoring"]

  analysis_config = {
    bucket_span = "15m"
    detectors = [
      {
        function             = "count"
        detector_description = "Count anomalies in web traffic"
      },
      {
        function             = "mean"
        field_name           = "response_time"
        detector_description = "Mean response time anomalies"
      }
    ]
    influencers = ["client_ip", "status_code"]
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }

  analysis_limits = {
    model_memory_limit = "100mb"
  }

  model_plot_config = {
    enabled = true
  }

  datafeed_config = {
    datafeed_id = "datafeed-example-anomaly-detector"
    indices     = ["web-logs-*"]
    query = jsonencode({
      match_all = {}
    })
    frequency   = "60s"
    query_delay = "60s"
  }

  model_snapshot_retention_days = 30
  results_retention_days        = 90
}
