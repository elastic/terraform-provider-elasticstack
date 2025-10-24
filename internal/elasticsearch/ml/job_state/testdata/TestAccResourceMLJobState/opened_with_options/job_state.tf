variable "job_id" {
  description = "The job ID for the ML job"
  type        = string
}

variable "force" {
  description = "Whether to force the job state change"
  type        = bool
  default     = true
}

variable "job_timeout" {
  description = "Timeout for the job state change"
  type        = string
  default     = "1m"
}

provider "elasticstack" {
  elasticsearch {}
}

# First create an ML anomaly detection job
resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id      = var.job_id
  description = "Test anomaly detection job for state management with options"

  analysis_config = {
    bucket_span = "15m"
    detectors = [
      {
        function             = "count"
        detector_description = "Count detector"
      }
    ]
  }

  analysis_limits = {
    model_memory_limit = "100mb"
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
}

# Then manage the state of that ML job with custom options
resource "elasticstack_elasticsearch_ml_job_state" "test" {
  job_id      = elasticstack_elasticsearch_ml_anomaly_detection_job.test.job_id
  state       = "opened"
  force       = var.force
  job_timeout = var.job_timeout
}