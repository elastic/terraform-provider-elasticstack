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

# Regression test for https://github.com/elastic/terraform-provider-elasticstack/issues/2400.
# The histogram_custom_indicator good/total from and to fields were float32 in the
# generated client. Values like 0.001 are not exactly representable in float32:
#   float64(float32(0.001)) = 0.0010000000474974513
# causing a "provider produced inconsistent result after apply" error.
resource "elasticstack_kibana_slo" "test_slo" {
  name        = var.name
  slo_id      = "id-${var.name}"
  description = "SLO to test float64 precision in histogram range fields"

  histogram_custom_indicator {
    index           = "my-index-${var.name}"
    timestamp_field = "@timestamp"

    good {
      field       = "duration"
      aggregation = "range"
      from        = 0.001
      to          = 1.0
    }

    total {
      field       = "duration"
      aggregation = "value_count"
    }
  }

  time_window {
    duration = "7d"
    type     = "rolling"
  }

  budgeting_method = "occurrences"

  objective {
    target = 0.999
  }

  depends_on = [elasticstack_elasticsearch_index.my_index]
}
