# Apply steps: string-form KQL (broad Kibana create compatibility), sync_field, enabled.
variable "name" {
  type = string
}

variable "enabled" {
  type = bool
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name                = "my-index-${var.name}"
  deletion_protection = false
}

resource "elasticstack_kibana_slo" "test_slo" {
  name        = var.name
  description = "string kql with sync_field and enabled"
  enabled     = var.enabled

  kql_custom_indicator {
    index           = "my-index-${var.name}"
    filter          = "service.name: test"
    good            = "latency < 300"
    total           = "*"
    timestamp_field = "custom_timestamp"
  }

  time_window {
    duration = "7d"
    type     = "rolling"
  }

  budgeting_method = "timeslices"

  objective {
    target           = 0.95
    timeslice_target = 0.95
    timeslice_window = "5m"
  }

  space_id = "default"

  settings {
    sync_delay = "5m"
    frequency  = "5m"
    sync_field = "@timestamp"
  }

  tags = ["acc-slo-align"]

  depends_on = [elasticstack_elasticsearch_index.my_index]
}
