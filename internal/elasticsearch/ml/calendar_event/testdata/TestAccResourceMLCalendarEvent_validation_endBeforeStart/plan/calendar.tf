variable "holder_calendar_id" {
  description = "Calendar id for holder resource"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_calendar" "holder" {
  calendar_id = var.holder_calendar_id
  description = "holder for event validation"
}

resource "elasticstack_elasticsearch_ml_calendar_event" "bad" {
  calendar_id = elasticstack_elasticsearch_ml_calendar.holder.calendar_id
  description = "bad window"
  start_time  = "2026-06-02T06:00:00Z"
  end_time    = "2026-06-02T00:00:00Z"
}
