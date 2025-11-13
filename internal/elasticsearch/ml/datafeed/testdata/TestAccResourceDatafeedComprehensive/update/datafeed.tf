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
  description = "Test job for updated datafeed"

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

# Updated datafeed configuration with modified values
resource "elasticstack_elasticsearch_ml_datafeed" "test" {
  datafeed_id = var.datafeed_id
  job_id      = elasticstack_elasticsearch_ml_anomaly_detection_job.test.job_id
  indices     = ["test-index-1-*", "test-index-2-*", "test-index-3-*"] # Added new index

  query = jsonencode({
    bool = {
      must = [
        {
          range = {
            "@timestamp" = {
              gte = "now-2h" # Changed from 1h to 2h
            }
          }
        },
        {
          term = {
            "status" = "updated" # Changed from "active" to "updated"
          }
        }
      ]
    }
  })

  # Remove aggregations since they can't be used with script_fields
  # Only use script_fields for the update test
  script_fields = jsonencode({
    triple_value = { # Changed from double_value to triple_value
      script = {
        source = "doc['value'].value * 3" # Changed multiplier from 2 to 3
      }
    }
    status_lower = { # Changed from status_upper to status_lower
      script = {
        source = "doc['status'].value.toLowerCase()" # Changed to toLowerCase
      }
    }
  })

  runtime_mappings = jsonencode({
    day_of_week = { # Changed from hour_of_day to day_of_week
      type = "long"
      script = {
        source = "emit(doc['@timestamp'].value.dayOfWeek)" # Changed script
      }
    }
  })

  scroll_size        = 1000   # Changed from 500 to 1000
  frequency          = "60s"  # Changed from 30s to 60s
  query_delay        = "120s" # Changed from 60s to 120s
  max_empty_searches = 20     # Changed from 10 to 20

  chunking_config = {
    mode      = "manual" # Keep same mode as original
    time_span = "2h"     # Changed from 1h to 2h
  }

  delayed_data_check_config = {
    enabled      = false # Changed from true to false
    check_window = "4h"  # Changed from 2h to 4h
  }

  indices_options = {
    expand_wildcards   = ["open"] # Changed from ["open", "closed"] to ["open"]
    ignore_unavailable = false    # Changed from true to false
    allow_no_indices   = true     # Changed from false to true
    ignore_throttled   = true     # Changed from false to true
  }

  depends_on = [elasticstack_elasticsearch_ml_anomaly_detection_job.test]
}