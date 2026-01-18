provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name                = "my-index-budgetingmethodfail"
  deletion_protection = false
}

resource "elasticstack_kibana_slo" "test_slo" {
  name        = "budgetingmethodfail"
  slo_id      = "id-budgetingmethodfail"
  description = "fully sick SLO"

  apm_latency_indicator {
    environment      = "production"
    service          = "my-service"
    transaction_type = "request"
    transaction_name = "GET /sup/dawg"
    index            = "my-index-budgetingmethodfail"
    threshold        = 500
  }

  time_window {
    duration = "7d"
    type     = "rolling"
  }

  budgeting_method = "supdawg"

  objective {
    target           = 0.999
    timeslice_target = 0.95
    timeslice_window = "5m"
  }

  depends_on = [elasticstack_elasticsearch_index.my_index]
}
