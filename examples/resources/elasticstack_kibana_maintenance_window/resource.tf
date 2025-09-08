provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_maintenance_window" "my_maintenance_window" {
  title   = "UPDATE TEST"
  enabled = true

  custom_schedule = {
    start    = "1993-01-01T05:00:00.200Z"
    duration = "12d"

    recurring = {
      every        = "21d"
      on_week_day  = ["MO", "+3TU", "-2FR"]
      on_month_day = [1, 2, 4, 6, 7]
      on_month     = [12]
    }
  }

  scope = {
    alerting = {
      kql = "_id: '1234'"
    }
  }
}
