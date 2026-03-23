variable "name" {
  description = "The name of the SLO"
  type        = string
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
  slo_id           = "id-${var.name}"
  name             = var.name
  description      = "SLO created by the legacy SDK provider"
  budgeting_method = "occurrences"

  time_window {
    duration = "7d"
    type     = "rolling"
  }

  objective {
    target = 0.99
  }

  kql_custom_indicator {
    index           = "my-index-${var.name}"
    timestamp_field = "@timestamp"
    filter          = "*"
    good            = "status_code < 500"
    total           = "*"
  }

  depends_on = [elasticstack_elasticsearch_index.my_index]
}
