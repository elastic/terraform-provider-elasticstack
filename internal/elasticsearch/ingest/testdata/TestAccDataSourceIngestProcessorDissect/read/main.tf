provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_dissect" "test" {
  field   = "message"
  pattern = "%%{clientip} %%{ident} %%{auth} [%%{@timestamp}] \"%%{verb} %%{request} HTTP/%%{httpversion}\" %%{status} %%{size}"
}
