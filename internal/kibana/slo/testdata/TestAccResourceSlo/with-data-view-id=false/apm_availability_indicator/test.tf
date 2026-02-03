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
  slo_id      = "id-${var.name}"
  description = "fully sick SLO"

  apm_availability_indicator {
    environment      = "production"
    service          = "my-service"
    transaction_type = "request"
    transaction_name = "GET /sup/dawg"
    index            = "my-index-${var.name}"
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

  space_id = "default"

  settings = {
    sync_delay = "5m"
    frequency  = "5m"
  }

  depends_on = [elasticstack_elasticsearch_index.my_index]
}
