variable "name" {
  type = string
}

variable "tags" {
  type    = list(string)
  default = null
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
  description = "tags coverage"
  tags        = var.tags

  metric_custom_indicator {
    index = "my-index-${var.name}"
    good {
      metrics {
        name        = "A"
        aggregation = "sum"
        field       = "processor.processed"
      }
      equation = "A"
    }
    total {
      metrics {
        name        = "A"
        aggregation = "sum"
        field       = "processor.accepted"
      }
      equation = "A"
    }
  }

  time_window {
    duration = "7d"
    type     = "rolling"
  }

  budgeting_method = "timeslices"

  objective {
    target           = 0.95
    timeslice_target = 0.95
    timeslice_window = "5m"
  }

  depends_on = [elasticstack_elasticsearch_index.my_index]
}
