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

# This config uses float64 literal values that are not exactly representable in
# float32 (e.g. 0.999). When the objective fields were float32 in the generated
# client, the provider corrupted these values on read:
#   float64(float32(0.999)) = 0.9990000128746033
# causing a "provider produced inconsistent result after apply" error.
# See https://github.com/elastic/terraform-provider-elasticstack/issues/2396.
resource "elasticstack_kibana_slo" "test_slo" {
  name        = var.name
  slo_id      = "id-${var.name}"
  description = "SLO to test float64 precision in objective fields"

  kql_custom_indicator {
    index           = "my-index-${var.name}"
    good            = "latency < 300"
    total           = "*"
    timestamp_field = "@timestamp"
  }

  time_window {
    duration = "7d"
    type     = "rolling"
  }

  budgeting_method = "timeslices"

  objective {
    target           = 0.999
    timeslice_target = 0.95
    timeslice_window = "5m"
  }

  depends_on = [elasticstack_elasticsearch_index.my_index]
}
