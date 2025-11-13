provider "elasticstack" {
  elasticsearch {}
}

# First create an ML anomaly detection job
resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "example" {
  job_id      = "example-ml-job"
  description = "Example anomaly detection job"

  analysis_config = {
    bucket_span = "15m"
    detectors = [
      {
        function             = "count"
        detector_description = "Count detector"
      }
    ]
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
}

# Manage the state of the ML job - open it
resource "elasticstack_elasticsearch_ml_job_state" "example" {
  job_id = elasticstack_elasticsearch_ml_anomaly_detection_job.example.job_id
  state  = "opened"

  # Optional settings
  force       = false
  job_timeout = "30s"

  # Timeouts for asynchronous operations
  timeouts = {
    create = "5m"
    update = "5m"
  }

  depends_on = [elasticstack_elasticsearch_ml_anomaly_detection_job.example]
}

# Example with different configuration options
resource "elasticstack_elasticsearch_ml_job_state" "example_with_options" {
  job_id = elasticstack_elasticsearch_ml_anomaly_detection_job.example.job_id
  state  = "closed"

  # Use force close for quicker shutdown
  force = true

  # Custom timeout
  job_timeout = "2m"

  # Custom timeouts for asynchronous operations  
  timeouts = {
    create = "10m"
    update = "3m"
  }

  depends_on = [elasticstack_elasticsearch_ml_anomaly_detection_job.example]
}