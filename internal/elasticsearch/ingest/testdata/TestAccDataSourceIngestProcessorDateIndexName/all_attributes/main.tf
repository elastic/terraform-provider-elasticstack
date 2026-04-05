provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_date_index_name" "test" {
  field              = "event_date"
  date_rounding      = "d"
  date_formats       = ["ISO8601", "UNIX"]
  index_name_prefix  = "logs-"
  timezone           = "America/New_York"
  locale             = "FRENCH"
  index_name_format  = "yyyy-MM"
  description        = "route documents by event date"
  if                 = "ctx.event_date != null"
  ignore_failure     = true
  tag                = "date-index-name-tag"
}
