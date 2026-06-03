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
  description = "SLO with artifacts dashboards"

  kql_custom_indicator {
    index           = "my-index-${var.name}"
    filter          = "*"
    good            = "latency < 300"
    total           = "*"
    timestamp_field = "@timestamp"
  }

  time_window {
    duration = "7d"
    type     = "rolling"
  }

  budgeting_method = "occurrences"

  objective {
    target = 0.99
  }

  artifacts = {
    dashboards = [
      {
        id = "test-dashboard-id-${var.name}"
      }
    ]
  }

  depends_on = [elasticstack_elasticsearch_index.my_index]
}
