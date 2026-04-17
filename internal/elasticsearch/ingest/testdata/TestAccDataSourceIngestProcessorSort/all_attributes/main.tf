provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_sort" "test" {
  field          = "items"
  order          = "desc"
  target_field   = "sorted_items"
  description    = "sort array"
  if             = "ctx.items != null"
  ignore_failure = true
  on_failure = [
    jsonencode({
      append = {
        field = "errors"
        value = "sort_failed"
      }
    })
  ]
  tag = "sort-items"
}
