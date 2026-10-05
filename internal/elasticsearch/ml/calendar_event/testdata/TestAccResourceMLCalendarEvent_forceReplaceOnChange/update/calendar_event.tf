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
  description = "Replace test changed description"
  start_time  = "2027-01-01T00:00:00Z"
  end_time    = "2027-01-01T01:00:00Z"
}
