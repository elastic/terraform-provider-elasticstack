provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_user_agent" "test" {
  field               = "http.request.headers.user-agent"
  target_field        = "user_agent_details"
  regex_file          = "custom-regexes.yml"
  properties          = ["os", "name", "device"]
  extract_device_type = true
  ignore_missing      = true
  ignore_failure      = true
  description         = "parse user agent"
  if                  = "ctx.agent != null"
  tag                 = "ua-tag"
  on_failure          = ["{\"set\":{\"field\":\"error.message\",\"value\":\"ua failed\"}}"]
}
