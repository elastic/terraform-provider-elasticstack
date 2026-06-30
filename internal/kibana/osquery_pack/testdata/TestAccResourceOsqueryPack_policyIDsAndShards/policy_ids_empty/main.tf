variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_osquery_pack" "test" {
  name    = "tf-acc-osquery-pack-policy-${var.suffix}"
  enabled = true

  queries = {
    find_procs = {
      query = "SELECT pid, name FROM processes LIMIT 5;"
    }
  }
}
