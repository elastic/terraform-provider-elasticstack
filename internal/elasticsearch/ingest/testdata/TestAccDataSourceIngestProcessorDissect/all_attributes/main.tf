provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_dissect" "test" {
  field            = "message"
  pattern          = "%%{clientip} %%{ident} %%{auth} [%%{@timestamp}] \"%%{verb} %%{request} HTTP/%%{httpversion}\" %%{status} %%{size}"
  append_separator = "|"
  ignore_missing   = true
  ignore_failure   = true
  description      = "Dissect log line"
  if               = "ctx.message != null"
  tag              = "dissect-tag"
}
