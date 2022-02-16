provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_grok" "grok" {
  field    = "message"
  patterns = ["%%{FAVORITE_DOG:pet}", "%%{FAVORITE_CAT:pet}"]
  pattern_definitions = {
    FAVORITE_DOG = "beagle"
    FAVORITE_CAT = "burmese"
  }
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "grok-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_grok.grok.json
  ]
}
