provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_script" "test" {
  description    = "Annotate tags when env is present"
  if             = "ctx.env != null"
  ignore_failure = true
  lang           = "expression"
  tag            = "script-tag"

  source = <<EOF
ctx['tag_count'] = params['count'];
ctx['tag_prefix'] = params['prefix'];
EOF

  params = jsonencode({
    count  = 2
    prefix = "prod"
  })

  on_failure = [
    jsonencode({
      set = {
        field = "error.message"
        value = "script processor failed"
      }
    })
  ]
}
