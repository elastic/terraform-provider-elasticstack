provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_script" "script" {
  description = "Extract 'tags' from 'env' field"
  lang        = "painless"

  source = <<EOF
String[] envSplit = ctx['env'].splitOnToken(params['delimiter']);
ArrayList tags = new ArrayList();
tags.add(envSplit[params['position']].trim());
ctx['tags'] = tags;
EOF

  params = jsonencode({
    delimiter = "-"
    position  = 1
  })

}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "script-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_script.script.json
  ]
}
