provider "elasticstack" {
  elasticsearch {
    username  = "elastic"
    password  = "password"
    endpoints = ["http://localhost:9200"]
  }
}

resource "elasticstack_kibana_slo" "test_slo" {
  name        = "%s"
  description = "my kewl SLO"
  indicator {
    type = "sli.apm.transactionDuration"
    params = {
      environment     = "production"
      service         = "my-service"
      transactionType = "request"
      transactionName = "GET /sup/dawg"
      index           = "my-index"
      threshold       = 500
    }
  }

  time_window {
    duration   = "1w"
    isCalendar = true
  }

  budgetingMethod = "timeslices"

  objective {
    target          = 0.999
    timesliceTarget = 0.95
    timesliceWindow = "5m"
  }

  settings {
    syncDelay = "5m"
    frequency = "1m"
  }
}
