provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_date" "test" {
  field          = "event_date"
  target_field   = "parsed_timestamp"
  formats        = ["ISO8601", "UNIX", "dd/MM/yyyy"]
  timezone       = "America/New_York"
  locale         = "FRENCH"
  output_format  = "yyyy-MM-dd"
  description    = "Parse date from event_date field"
  if             = "ctx.event_date != null"
  ignore_failure = true
  tag            = "date-tag"
}
