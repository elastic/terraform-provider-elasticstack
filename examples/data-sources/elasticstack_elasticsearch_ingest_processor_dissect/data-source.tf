provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_dissect" "dissect" {
  field   = "message"
  pattern = "%%{clientip} %%{ident} %%{auth} [%%{@timestamp}] \"%%{verb} %%{request} HTTP/%%{httpversion}\" %%{status} %%{size}"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "dissect-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_dissect.dissect.json
  ]
}
