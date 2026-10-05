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
  calendar_id      = elasticstack_elasticsearch_ml_calendar.test.calendar_id
  description      = "Unsupported version scheduling fields test"
  start_time       = "2027-02-01T00:00:00Z"
  end_time         = "2027-02-01T02:00:00Z"
  force_time_shift = "3600"
}
