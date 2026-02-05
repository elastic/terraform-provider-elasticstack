variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_slo" "xp_upjet_ext_api_duration" {
  name        = "[Crossplane] Managed Resource External API Request Duration ${var.suffix}"
  description = "Measures in seconds how long it takes a Cloud SDK call to complete. This measures the time it takes for Crossplane Provider pods to complete external API requests, for example, the provider-aws-ec2 talks to AWS' EC2 API."
  slo_id      = "xp_upjet_ext_api_duration_${var.suffix}"

  histogram_custom_indicator {
    index = "metrics-*:metrics-*"
    good {
      field       = "prometheus.upjet_resource_ext_api_duration.histogram"
      aggregation = "range"
      from        = 0
      # 10s
      to = 10
    }
    total {
      field       = "prometheus.upjet_resource_ext_api_duration.histogram"
      aggregation = "range"
      from        = 0
      to          = 999999
    }
    filter          = "prometheus.upjet_resource_ext_api_duration.histogram: *"
    timestamp_field = "@timestamp"
  }

  time_window {
    duration = "30d"
    type     = "rolling"
  }

  budgeting_method = "occurrences"

  objective {
    target = 0.99
  }

  group_by = ["orchestrator.cluster.name"]

  tags = ["crossplane", "infra-mki"]
}
