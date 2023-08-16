provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_slo" "auth_server_latency" {
  name        = "Auth server latency"
  description = "Ensures auth server is responding in time"

  apm_latency_indicator {
    environment      = "production"
    service          = "auth"
    transaction_type = "request"
    transaction_name = "GET /auth"
    index            = "metrics-apm*"
    threshold        = 500
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

  settings {
    sync_delay = "5m"
    frequency  = "5m"
  }

}

resource "elasticstack_kibana_slo" "auth_server_availability" {
  name        = "Auth server latency"
  description = "Ensures auth server is responding in time"

  apm_availability_indicator {
    environment      = "production"
    service          = "my-service"
    transaction_type = "request"
    transaction_name = "GET /sup/dawg"
    index            = "metrics-apm*"
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

  settings {
    sync_delay = "5m"
    frequency  = "5m"
  }

}

resource "elasticstack_kibana_slo" "custom_kql" {
  name        = "custom kql"
  description = "custom kql"

  kql_custom_indicator {
    index           = "my-index"
    good            = "latency < 300"
    total           = "*"
    filter          = "labels.groupId: group-0"
    timestamp_field = "custom_timestamp"
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

  settings {
    sync_delay = "5m"
    frequency  = "5m"
  }

}

//Available from 8.10.0
resource "elasticstack_kibana_slo" "custom_histogram" {
  name        = "custom histogram"
  description = "custom histogram"

  histogram_custom_indicator {
    index = "my-index"
    good {
      field       = "test"
      aggregation = "value_count"
      filter      = "latency < 300"
    }
    total {
      field       = "test"
      aggregation = "value_count"
    }
    filter          = "labels.groupId: group-0"
    timestamp_field = "custom_timestamp"
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

}

//Available from 8.10.0
resource "elasticstack_kibana_slo" "custom_metric" {
  name        = "custom kql"
  description = "custom kql"

  metric_custom_indicator {
    index = "my-index"
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

}
