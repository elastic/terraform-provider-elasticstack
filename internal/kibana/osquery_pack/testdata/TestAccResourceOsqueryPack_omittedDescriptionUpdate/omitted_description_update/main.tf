variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_osquery_pack" "test" {
  name    = "tf-acc-osquery-pack-no-desc-${var.suffix}"
  enabled = false

  queries = {
    find_procs = {
      query = "SELECT pid, name, path FROM processes LIMIT 10;"
    }
  }
}
