variable "holder_calendar_id" {
  description = "Calendar id for holder resource"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_calendar" "holder" {
  calendar_id = var.holder_calendar_id
  description = "holder for malformed start_time validation"
}

resource "elasticstack_elasticsearch_ml_calendar_event" "bad" {
  calendar_id = elasticstack_elasticsearch_ml_calendar.holder.calendar_id
  description = "malformed start_time"
  start_time  = "not-a-time"
  end_time    = "2026-06-01T01:00:00Z"
}
