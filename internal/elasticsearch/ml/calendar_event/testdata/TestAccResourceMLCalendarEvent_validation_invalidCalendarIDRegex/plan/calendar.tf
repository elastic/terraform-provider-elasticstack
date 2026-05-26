variable "holder_calendar_id" {
  description = "Calendar id for holder resource"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_calendar" "holder" {
  calendar_id = var.holder_calendar_id
  description = "holder"
}

resource "elasticstack_elasticsearch_ml_calendar_event" "bad" {
  calendar_id = "INVALID_EVENT_CAL"
  description = "x"
  start_time  = "2026-06-01T00:00:00Z"
  end_time    = "2026-06-01T01:00:00Z"
  depends_on  = [elasticstack_elasticsearch_ml_calendar.holder]
}
