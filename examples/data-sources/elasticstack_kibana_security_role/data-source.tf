provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

# Manage a role and then look it up with the data source. The data source's
# read is deferred until after the role is created (via depends_on), so the
# lookup never targets a role that does not yet exist.
resource "elasticstack_kibana_security_role" "example" {
  name = "sample_role"

  elasticsearch {
    cluster = ["manage_ingest_pipelines", "manage_index_templates", "auto_configure"]

    indices {
      names      = ["logs-myapp-*"]
      privileges = ["write", "create_index", "auto_configure"]
    }
  }
}

data "elasticstack_kibana_security_role" "example" {
  name = elasticstack_kibana_security_role.example.name

  depends_on = [elasticstack_kibana_security_role.example]
}
