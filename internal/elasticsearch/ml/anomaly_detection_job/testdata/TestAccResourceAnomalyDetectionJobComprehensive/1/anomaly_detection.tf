variable "job_id" {
  description = "The job ID for the anomaly detection job"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id      = var.job_id
  description = "Comprehensive test anomaly detection job"
  groups      = ["test-group", "ml-group"]

  analysis_config = {
    bucket_span = "10m"
    detectors = [
      {
        function              = "count"
        detector_description = "Count detector"
      },
      {
        function              = "mean"
        field_name           = "response_time"
        detector_description = "Mean response time detector"
      }
    ]
    influencers = ["status_code"]
  }

  analysis_limits = {
    model_memory_limit = "100mb"
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }

  model_plot_config = {
    enabled = true
  }

  allow_lazy_open = true
  background_persist_interval = "1h"
  custom_settings = "{\"custom_key\": \"custom_value\"}"
  daily_model_snapshot_retention_after_days = 3
  model_snapshot_retention_days = 7
  renormalization_window_days = 14
  results_retention_days = 30
}