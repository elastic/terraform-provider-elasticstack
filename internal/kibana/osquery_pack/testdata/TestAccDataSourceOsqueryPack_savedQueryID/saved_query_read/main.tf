variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_osquery_saved_query" "test" {
  saved_query_id = "tf-acc-osquery-sq-ds-${var.suffix}"
  query          = "SELECT pid, name FROM processes LIMIT 5;"
  interval       = 3600
}

resource "elasticstack_kibana_osquery_pack" "test" {
  name    = "tf-acc-osquery-pack-sq-ds-${var.suffix}"
  enabled = true

  queries = {
    find_procs = {
      query          = "SELECT pid, name FROM processes LIMIT 5;"
      saved_query_id = elasticstack_kibana_osquery_saved_query.test.saved_query_id
    }
  }
}

data "elasticstack_kibana_osquery_pack" "test" {
  pack_id = elasticstack_kibana_osquery_pack.test.pack_id
}
