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
  calendar_id       = elasticstack_elasticsearch_ml_calendar.test.calendar_id
  description       = "ACC outage with optional scheduling fields"
  start_time        = "2026-09-01T00:00:00Z"
  end_time          = "2026-09-01T02:00:00Z"
  skip_result       = true
  skip_model_update = true
  force_time_shift  = "3600"
}
