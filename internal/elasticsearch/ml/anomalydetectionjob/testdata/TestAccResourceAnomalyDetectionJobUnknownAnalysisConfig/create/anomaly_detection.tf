# Regression test for #3403: plan must not fail with a Value Conversion Error when
# analysis_config is sourced from an unknown value (e.g. each.value.job.analysis_config in for_each).
#
# terraform_data.source.output is unknown during the first plan (before apply),
# reproducing the for_each pattern.
variable "job_id" {
  description = "The job ID for the anomaly detection job"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "terraform_data" "source" {
  input = {
    bucket_span = "15m"
    detectors = [
      {
        function = "count"
      }
    ]
    influencers            = []
    categorization_filters = []
  }
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id          = var.job_id
  analysis_config = terraform_data.source.output

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
}
