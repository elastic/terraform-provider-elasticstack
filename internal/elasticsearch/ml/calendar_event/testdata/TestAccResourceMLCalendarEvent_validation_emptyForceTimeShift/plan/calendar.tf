variable "holder_calendar_id" {
  description = "Calendar id for holder resource"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_calendar" "holder" {
  calendar_id = var.holder_calendar_id
  description = "holder for force_time_shift validation"
}

resource "elasticstack_elasticsearch_ml_calendar_event" "bad" {
  calendar_id      = elasticstack_elasticsearch_ml_calendar.holder.calendar_id
  description      = "empty force_time_shift"
  start_time       = "2026-06-01T00:00:00Z"
  end_time         = "2026-06-01T01:00:00Z"
  force_time_shift = ""
}
