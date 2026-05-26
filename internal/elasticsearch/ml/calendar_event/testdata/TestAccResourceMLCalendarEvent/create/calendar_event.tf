variable "calendar_id" {
  description = "The calendar ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_calendar" "test" {
  calendar_id = var.calendar_id
}

resource "elasticstack_elasticsearch_ml_calendar_event" "test" {
  calendar_id = elasticstack_elasticsearch_ml_calendar.test.calendar_id
  description = "Planned maintenance"
  start_time  = "2026-06-01T00:00:00Z"
  end_time    = "2026-06-01T06:00:00Z"
}
