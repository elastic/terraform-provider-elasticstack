provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "remote" {
  name                 = "Remote Elasticsearch"
  output_id            = "remote-es-output"
  type                 = "remote_elasticsearch"
  service_token        = var.remote_service_token
  default_integrations = false
  default_monitoring   = false

  hosts = [
    "https://remote-elasticsearch.example.com:9200",
  ]

  # Optional: automatic integration asset sync to the remote cluster (subscription/version limits apply).
  sync_integrations             = true
  sync_uninstalled_integrations = false
  write_to_logs_streams         = false

  ssl = {
    certificate_authorities = [file("${path.module}/remote-ca.pem")]
  }
}

variable "remote_service_token" {
  type        = string
  sensitive   = true
  description = "Service token for the remote Elasticsearch cluster (create per Elastic/Fleet docs)."
}
