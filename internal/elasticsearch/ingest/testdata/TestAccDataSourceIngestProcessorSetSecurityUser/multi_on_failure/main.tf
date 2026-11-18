provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_set_security_user" "test" {
  field = "user"
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "set security user failed"
      }
    }),
    jsonencode({
      set = {
        field = "error.type"
        value = "set_security_user_error"
      }
    }),
  ]
}
