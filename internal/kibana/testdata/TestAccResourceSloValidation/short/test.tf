variable "name" {
  type = string
}

variable "slo_id" {
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
  slo_id      = var.slo_id
  description = "fully sick SLO"

  apm_latency_indicator {
    environment      = "production"
    service          = "my-service"
    transaction_type = "request"
    transaction_name = "GET /sup/dawg"
    index            = "my-index-${var.name}"
    threshold        = 500
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
