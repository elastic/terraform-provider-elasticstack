variable "job_id" {
  type = string
}

variable "filter_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_filter" "scope_test" {
  filter_id   = var.filter_id
  description = "Acceptance test filter for custom_rules scope"
  items       = ["10.0.0.1"]
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id = var.job_id

  depends_on = [elasticstack_elasticsearch_ml_filter.scope_test]

  analysis_config = {
    bucket_span = "3h"
    detectors = [
      {
        function        = "count"
        over_field_name = "clientip"
        custom_rules = [
          {
            actions = ["skip_result"]
            scope = {
              clientip = {
                filter_id   = var.filter_id
                filter_type = "exclude"
              }
            }
          }
        ]
      }
    ]
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
}
