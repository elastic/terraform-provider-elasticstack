provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_slo" "fleetctl_api_pod_readiness" {
  name        = var.name
  description = "API Pod is running"

  kql_custom_indicator {
    index           = "metrics-*,serverless-metrics-*:metrics-*"
    good            = "kubernetes.pod.status.ready: true"
    total           = "*"
    timestamp_field = "@timestamp"
  }

  time_window {
    duration = "7d"
    type     = "rolling"
  }

  budgeting_method = "occurrences"

  objective {
    target = 0.9
  }

  settings {
    sync_delay = "1m"
    frequency  = "1m"
  }

  tags = var.tags
}
