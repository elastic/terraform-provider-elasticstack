provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_grok" "test" {
  field    = "message"
  patterns = ["%%{FAVORITE_DOG:pet}", "%%{FAVORITE_CAT:pet}"]
  pattern_definitions = {
    FAVORITE_DOG = "beagle"
    FAVORITE_CAT = "burmese"
  }
}
