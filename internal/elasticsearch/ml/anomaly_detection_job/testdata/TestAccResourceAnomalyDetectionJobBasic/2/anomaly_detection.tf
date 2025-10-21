variable "job_id" {
  description = "The job ID for the anomaly detection job"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id      = var.job_id
  description = "Updated basic test anomaly detection job"

  analysis_config = {
    bucket_span = "15m"
    detectors = [
      {
        function              = "count"
        detector_description = "Count detector"
      }
    ]
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }

  groups = ["basic-group"]
  
  analysis_limits = {
    model_memory_limit = "128mb"
  }
  
  allow_lazy_open = true
  results_retention_days = 15
}