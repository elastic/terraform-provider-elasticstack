provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_csv" "csv" {
  field         = "my_field"
  target_fields = ["field1", "field2"]
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "csv-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_csv.csv.json
  ]
}
