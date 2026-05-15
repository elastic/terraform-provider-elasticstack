# Regression test for #2966: plan must not fail with a Value Conversion Error when
# analysis_config.detectors is sourced from a Terraform variable.
variable "job_id" {
  description = "The job ID for the anomaly detection job"
  type        = string
}

variable "detectors" {
  description = "Detector configuration sourced from a variable (reproduces #2966 regression)"
  type = list(object({
    function             = string
    field_name           = optional(string)
    by_field_name        = optional(string)
    detector_description = optional(string)
  }))
  default = [
    {
      function = "count"
    }
  ]
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id = var.job_id

  analysis_config = {
    bucket_span = "15m"
    detectors   = var.detectors
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
}
