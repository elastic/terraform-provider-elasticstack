variable "name" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name                = "my-index"
  deletion_protection = false
}

resource "elasticstack_kibana_slo" "test_slo" {
  name        = var.name
  description = "doc_count metric custom indicator"

  metric_custom_indicator {
    index           = "my-index"
    timestamp_field = "@timestamp"

    good {
      metrics {
        name        = "A"
        aggregation = "doc_count"
        filter      = "status: success"
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

  budgeting_method = "occurrences"

  objective {
    target = 0.95
  }

  time_window {
    duration = "7d"
    type     = "rolling"
  }

  depends_on = [elasticstack_elasticsearch_index.my_index]
}
