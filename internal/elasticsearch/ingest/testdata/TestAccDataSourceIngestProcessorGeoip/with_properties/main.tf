provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_geoip" "test" {
  field      = "ip"
  properties = ["ip", "country_name", "city_name"]
}
