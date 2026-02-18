variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_slo" "test_slo" {
  name        = var.name
  description = "fleetctl-like configuration with empty total"

  kql_custom_indicator {
    index  = "metrics-*"
    filter = "kubernetes.deployment.name: \"fleetctl-api\" and kubernetes.pod.status.ready : *"
    good   = "kubernetes.pod.status.ready: true"
    # total is not specified, should default to ""
  }

  budgeting_method = "occurrences"

  objective {
    target = 0.95
  }

  time_window {
    duration = "7d"
    type     = "rolling"
  }
}
