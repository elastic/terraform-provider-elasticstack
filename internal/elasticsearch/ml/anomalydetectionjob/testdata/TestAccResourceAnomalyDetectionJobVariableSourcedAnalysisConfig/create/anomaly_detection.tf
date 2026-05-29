# Regression test for #3403: plan must not fail with a Value Conversion Error when
# analysis_config is sourced from a Terraform variable.
variable "job_id" {
  description = "The job ID for the anomaly detection job"
  type        = string
}

variable "analysis_config" {
  description = "Analysis configuration sourced from a variable (reproduces #3403 regression)"
  type = object({
    bucket_span               = optional(string)
    categorization_field_name = optional(string)
    categorization_filters    = optional(list(string))
    detectors = list(object({
      function             = string
      field_name           = optional(string)
      by_field_name        = optional(string)
      over_field_name      = optional(string)
      partition_field_name = optional(string)
      detector_description = optional(string)
      exclude_frequent     = optional(string)
      use_null             = optional(bool)
    }))
    influencers            = optional(list(string))
    latency                = optional(string)
    model_prune_window     = optional(string)
    multivariate_by_fields = optional(bool)
    per_partition_categorization = optional(object({
      enabled      = optional(bool)
      stop_on_warn = optional(bool)
    }))
    summary_count_field_name = optional(string)
  })
  default = {
    bucket_span = "15m"
    detectors = [
      {
        function = "count"
      }
    ]
  }
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id = var.job_id

  analysis_config = var.analysis_config

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
}
