variable "job_id" {
  description = "The ML job ID"
  type        = string
}

variable "datafeed_id" {
  description = "The ML datafeed ID"
  type        = string
}

variable "index_name" {
  description = "The index name"
  type        = string
}

variable "endpoints" {
  type = list(string)
}

variable "api_key" {
  type    = string
  default = ""
}

variable "username" {
  type    = string
  default = ""
}

variable "password" {
  type      = string
  sensitive = true
  default   = ""
}

provider "elasticstack" {
  elasticsearch {}
}

# The anomaly job and its infrastructure use the default provider connection.
# The datafeed_state resource uses an explicit elasticsearch_connection block
# to exercise the scoped client code path.
resource "elasticstack_elasticsearch_index" "test" {
  name                = var.index_name
  deletion_protection = false
  mappings = jsonencode({
    properties = {
      "@timestamp" = {
        type = "date"
      }
      value = {
        type = "double"
      }
    }
  })
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id      = var.job_id
  description = "Test job for datafeed state explicit connection testing"
  analysis_config = {
    bucket_span = "1h"
    detectors = [{
      function             = "count"
      detector_description = "count"
    }]
  }
  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
  analysis_limits = {
    model_memory_limit = "10mb"
  }
}

resource "elasticstack_elasticsearch_ml_job_state" "test" {
  job_id = elasticstack_elasticsearch_ml_anomaly_detection_job.test.job_id
  state  = "opened"

  lifecycle {
    ignore_changes = ["state"]
  }
}

resource "elasticstack_elasticsearch_ml_datafeed" "test" {
  datafeed_id = var.datafeed_id
  job_id      = elasticstack_elasticsearch_ml_anomaly_detection_job.test.job_id
  indices     = [elasticstack_elasticsearch_index.test.name]
  query = jsonencode({
    match_all = {}
  })
}

# Use state="stopped" so the datafeed is never started and never auto-stops.
# This makes the test deterministic: the API state always matches the config
# state, so ImportStateVerify succeeds without needing to ignore "state".
resource "elasticstack_elasticsearch_ml_datafeed_state" "test" {
  datafeed_id = elasticstack_elasticsearch_ml_datafeed.test.datafeed_id
  state       = "stopped"

  elasticsearch_connection {
    endpoints = var.endpoints
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.api_key == "" ? var.username : null
    password  = var.api_key == "" ? var.password : null
    insecure  = true
  }

  depends_on = [
    elasticstack_elasticsearch_ml_datafeed.test,
    elasticstack_elasticsearch_ml_job_state.test
  ]
}
