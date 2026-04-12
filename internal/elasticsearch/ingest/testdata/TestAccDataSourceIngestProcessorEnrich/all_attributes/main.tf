provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_enrich" "test" {
  field          = "email"
  target_field   = "user.profile"
  policy_name    = "users-policy"
  ignore_missing = true
  override       = false
  max_matches    = 3
  shape_relation = "INTERSECTS"
  description    = "Enrich user details from a policy"
  if             = "ctx.email != null"
  ignore_failure = true
  tag            = "enrich-users"
}
