provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "remote" {
  name                 = "Remote Elasticsearch"
  output_id            = "remote-es-output"
  type                 = "remote_elasticsearch"
  service_token        = "REPLACE_ME_REMOTE_CLUSTER_SERVICE_TOKEN"
  default_integrations = false
  default_monitoring   = false

  hosts = [
    "https://remote-elasticsearch.example.com:9200",
  ]

  ssl = {
    certificate_authorities = [trimspace(<<-EOT
      -----BEGIN CERTIFICATE-----
      MIIBkTCB+wIJAKHHCgV4Jh0FMA0GCSqGGSIwDAQEKBQYwOzEbMBkGA1UEAxMSRWxhc3RpY
      -----END CERTIFICATE-----
    EOT
    )]
  }
}
