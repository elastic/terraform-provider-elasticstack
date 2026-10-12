variable "space_id" {
  description = "The Kibana space ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = var.space_id
}

resource "elasticstack_kibana_maintenance_window" "test_maintenance_window" {
  title    = "Terraform Maintenance Window In Space"
  space_id = elasticstack_kibana_space.test.space_id

  custom_schedule = {
    start    = "1992-01-01T05:00:00.200Z"
    duration = "10d"
    timezone = "UTC"

    recurring = {
      every       = "20d"
      end         = "2029-05-17T05:05:00.000Z"
      on_week_day = ["MO", "TU"]
    }
  }
}
