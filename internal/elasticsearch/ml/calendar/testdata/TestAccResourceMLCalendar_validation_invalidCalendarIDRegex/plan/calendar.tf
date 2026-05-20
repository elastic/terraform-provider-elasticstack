provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_calendar" "bad" {
  calendar_id = "INVALID_UPPER"
  description = "x"
}
