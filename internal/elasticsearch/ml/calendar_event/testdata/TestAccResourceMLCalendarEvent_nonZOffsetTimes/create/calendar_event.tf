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
  description = "Non-Z UTC offset test"
  start_time  = "2027-03-01T02:00:00+02:00"
  end_time    = "2027-03-01T04:30:00+02:00"
}
