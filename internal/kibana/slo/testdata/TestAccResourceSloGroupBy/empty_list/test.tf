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
  slo_id      = "id-${var.name}"
  description = "SLO with empty group_by list"

  # Mirrors the reported issue: kql custom indicator + occurrences + group_by = []
  kql_custom_indicator {
    index           = "my-index-${var.name}"
    good            = "elasticsearch.health.overall.status:\"green\" or elasticsearch.health.overall.status:\"yellow\""
    total           = "elasticsearch.health.overall.status:\"green\" or elasticsearch.health.overall.status:\"yellow\" or elasticsearch.health.overall.status:\"red\""
    filter          = "log.logger:\"org.elasticsearch.health.HealthPeriodicLogger\""
    timestamp_field = "@timestamp"
  }

  group_by = []

  time_window {
    duration = "7d"
    type     = "rolling"
  }

  budgeting_method = "occurrences"

  objective {
    target = 0.9995
  }

  depends_on = [elasticstack_elasticsearch_index.my_index]
}
