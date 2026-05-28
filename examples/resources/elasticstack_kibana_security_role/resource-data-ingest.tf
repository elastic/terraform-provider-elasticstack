provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "data_ingest" {
  name = "data_ingest"

  elasticsearch {
    cluster = ["manage_ingest_pipelines", "manage_index_templates", "auto_configure"]

    indices {
      names      = ["logs-myapp-*"]
      privileges = ["write", "create_index", "auto_configure"]
    }
  }
}
