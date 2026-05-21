variable "calendar_id" {
  description = "Over-long calendar id for validation test"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_calendar" "bad" {
  calendar_id = var.calendar_id
  description = "x"
}
