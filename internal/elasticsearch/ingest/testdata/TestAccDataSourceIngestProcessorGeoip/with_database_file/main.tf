provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_geoip" "test" {
  field         = "ip"
  database_file = "GeoLite2-City.mmdb"
}
