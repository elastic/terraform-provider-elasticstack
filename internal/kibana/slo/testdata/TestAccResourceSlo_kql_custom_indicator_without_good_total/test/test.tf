variable "name" {
  type = string
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
  description = "kql indicator without good/total"

  kql_custom_indicator {
    index           = "my-index-${var.name}"
    timestamp_field = "custom_timestamp"
  }

  budgeting_method = "timeslices"

  objective {
    target           = 0.95
    timeslice_target = 0.95
    timeslice_window = "5m"
  }

  time_window {
    duration = "7d"
    type     = "rolling"
  }

  depends_on = [elasticstack_elasticsearch_index.my_index]
}

