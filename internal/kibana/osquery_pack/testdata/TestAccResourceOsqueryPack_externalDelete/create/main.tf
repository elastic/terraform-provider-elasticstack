variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_osquery_pack" "test" {
  name        = "tf-acc-osquery-pack-del-${var.suffix}"
  description = "External delete acceptance test pack"

  queries = {
    simple = {
      query = "SELECT 1;"
    }
  }
}
