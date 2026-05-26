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
  description       = "False scheduling flags test"
  start_time        = "2026-11-01T00:00:00Z"
  end_time          = "2026-11-01T02:00:00Z"
  skip_result       = false
  skip_model_update = false
  force_time_shift  = "7200"
}
