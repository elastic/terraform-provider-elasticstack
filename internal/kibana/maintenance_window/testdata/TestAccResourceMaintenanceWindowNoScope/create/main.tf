provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_maintenance_window" "test_maintenance_window" {
  title = "Terraform Maintenance Window NTH DAY"

  custom_schedule = {
    start    = "1999-02-02T05:00:00.200Z"
    duration = "1d"
    timezone = "Asia/Taipei"

    recurring = {
      every       = "1w"
      occurrences = 5
      on_week_day = ["+1MO", "-2FR"]
    }
  }
}
