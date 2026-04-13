provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_grok" "test" {
  description       = "Parse ECS-compatible log lines"
  ecs_compatibility = "v1"
  field             = "log.original"
  patterns = [
    "%%{CUSTOMLEVEL:log.level} %%{GREEDYDATA:message}",
    "%%{CUSTOMLEVEL:log.level}",
  ]
  pattern_definitions = {
    CUSTOMLEVEL = "INFO|WARN|ERROR"
  }
  trace_match    = true
  ignore_missing = true
  if             = "ctx.log != null"
  ignore_failure = true
  tag            = "grok-all-attributes"
}
