provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_script" "test" {
  source = "ctx.result = 'ok';"

  params = jsonencode({
    tags = ["a", "b"]
    meta = {
      k = "v"
    }
  })
}
