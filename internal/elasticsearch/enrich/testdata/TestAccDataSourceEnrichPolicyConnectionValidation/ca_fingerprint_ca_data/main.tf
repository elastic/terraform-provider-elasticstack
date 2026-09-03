provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_enrich_policy" "test" {
  name = "validation"

  elasticsearch_connection {
    ca_fingerprint = "aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899"
    ca_data        = "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----"
  }
}
