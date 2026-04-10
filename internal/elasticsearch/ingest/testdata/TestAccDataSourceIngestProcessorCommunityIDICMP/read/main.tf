provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_community_id" "test" {
  transport = "icmp"
  icmp_type = 3
  icmp_code = 1
}
