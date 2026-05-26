# Acceptance exercises the scoped Elasticsearch client path: when
# elasticsearch_connection is set, ProviderClientFactory.GetElasticsearchClient
# builds a new typed client from that block instead of reusing the provider
# default (see internal/clients/provider_client_factory.go).
#
# The default (un-aliased) provider configures Elasticsearch with unreachable
# endpoints and invalid credentials so that real API calls would fail if this
# resource incorrectly fell back to the provider default client instead of the
# per-resource connection.
#
# The anomaly detection job uses provider.elasticstack.setup with elasticsearch {}
# so it still inherits acceptance credentials from the environment (same pattern
# as other ML tests that need a working job alongside a scoped resource).
#
# Note: ELASTICSEARCH_* environment variables may still override some connection
# fields when building clients (internal/clients/config/elasticsearch.go and
# base.go withEnvironmentOverrides). CI usually sets those variables; the invalid
# default provider block still documents intent and protects runs without those
# overrides. ImportState for this resource uses an empty connection list and
# therefore the provider default client, so import is covered by
# TestAccResourceMLCalendarJob_import instead of this test.

variable "calendar_id" {
  type = string
}

variable "job_id" {
  type = string
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
  type    = string
  default = ""
}

provider "elasticstack" {
  elasticsearch {
    endpoints = ["http://127.0.0.1:59997"]
    username  = "__elasticstack_ml_calendar_job_explicit_conn_invalid_default__"
    password  = "__elasticstack_ml_calendar_job_explicit_conn_invalid_default__"
  }
}

provider "elasticstack" {
  alias = "setup"

  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "job" {
  provider = elasticstack.setup

  job_id      = var.job_id
  description = "ACC job for ml_calendar_job explicit ES connection"

  analysis_config = {
    bucket_span = "15m"
    detectors = [
      {
        function = "count"
      }
    ]
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }
}

resource "elasticstack_elasticsearch_ml_calendar_job" "test" {
  calendar_id = var.calendar_id
  job_id      = elasticstack_elasticsearch_ml_anomaly_detection_job.job.job_id

  elasticsearch_connection {
    endpoints = var.endpoints
    api_key   = var.api_key != "" ? var.api_key : null
    username  = var.api_key == "" ? var.username : null
    password  = var.api_key == "" ? var.password : null
    insecure  = true
  }
}
