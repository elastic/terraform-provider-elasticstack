provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_set_security_user" "test" {
  field          = "actor"
  properties     = ["username", "roles", "email"]
  description    = "set security user metadata"
  if             = "ctx.user != null"
  ignore_failure = true
  tag            = "set-security-user"
  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "fallback"
      }
    })
  ]
}
