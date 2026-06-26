provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

# Placeholder resource shell for import-only testing; this step never runs apply.
resource "elasticstack_kibana_osquery_pack" "test" {
  name = "import-placeholder"

  queries = {
    q = {
      query = "SELECT 1;"
    }
  }
}
