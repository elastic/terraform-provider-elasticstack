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
  description = "doc_count metric custom indicator SLO"

  metric_custom_indicator {
    index = "my-index-${var.name}"

    good {
      metrics {
        name        = "A"
        aggregation = "doc_count"
        filter      = "status: 200"
      }
      equation = "A"
    }

    total {
      metrics {
        name        = "B"
        aggregation = "doc_count"
      }
      equation = "B"
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
