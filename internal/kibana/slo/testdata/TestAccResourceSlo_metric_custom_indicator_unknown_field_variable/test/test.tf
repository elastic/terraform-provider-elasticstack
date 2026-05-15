variable "name" {
  type = string
}

variable "metric_field_suffix" {
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
  description = "metric custom indicator unknown field variable regression test"

  metric_custom_indicator {
    index = "my-index-${var.name}"

    good {
      metrics {
        name        = "A"
        aggregation = "sum"
        field       = "processor.${var.metric_field_suffix}"
      }
      equation = "A"
    }

    total {
      metrics {
        name        = "A"
        aggregation = "sum"
        field       = "processor.${var.metric_field_suffix}"
      }
      equation = "A"
    }
  }

  time_window {
    duration = "7d"
    type     = "rolling"
  }

  budgeting_method = "occurrences"

  objective {
    target = 0.99
  }

  depends_on = [elasticstack_elasticsearch_index.my_index]
}
