variable "job_id" {
  type = string
}

variable "filter_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

# ML filter is created out-of-band by the test (Elasticsearch PUT _ml/filters/{filter_id}).

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id = var.job_id

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
                filter_type = "include"
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

  analysis_limits = {
    model_memory_limit = "10mb"
  }

  allow_lazy_open = true
}
