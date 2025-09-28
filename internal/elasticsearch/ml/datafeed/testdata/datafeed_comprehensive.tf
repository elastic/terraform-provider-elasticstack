variable "job_id" {
  description = "The ML job ID"
  type        = string
}

variable "datafeed_id" {
  description = "The ML datafeed ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

# First create the ML job
resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id      = var.job_id
  description = "Test job for comprehensive datafeed"

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

# Then create the comprehensive datafeed with all available attributes
resource "elasticstack_elasticsearch_ml_datafeed" "test" {
  datafeed_id = var.datafeed_id
  job_id      = elasticstack_elasticsearch_ml_anomaly_detection_job.test.job_id
  indices     = ["test-index-1-*", "test-index-2-*"]

  query = jsonencode({
    bool = {
      must = [
        {
          range = {
            "@timestamp" = {
              gte = "now-1h"
            }
          }
        },
        {
          term = {
            "status" = "active"
          }
        }
      ]
    }
  })

  # Remove aggregations and script_fields since they can't be used together
  # Only include script_fields for this comprehensive test
  script_fields = jsonencode({
    double_value = {
      script = {
        source = "doc['value'].value * 2"
      }
    }
    status_upper = {
      script = {
        source = "doc['status'].value.toUpperCase()"
      }
    }
  })

  runtime_mappings = jsonencode({
    hour_of_day = {
      type = "long"
      script = {
        source = "emit(doc['@timestamp'].value.hour)"
      }
    }
  })

  scroll_size        = 500
  frequency          = "30s"
  query_delay        = "60s"
  max_empty_searches = 10

  chunking_config = {
    mode      = "manual"
    time_span = "1h"
  }

  delayed_data_check_config = {
    enabled      = true
    check_window = "2h"
  }

  indices_options = {
    expand_wildcards   = ["open", "closed"]
    ignore_unavailable = true
    allow_no_indices   = false
    ignore_throttled   = false
  }

  depends_on = [elasticstack_elasticsearch_ml_anomaly_detection_job.test]
}