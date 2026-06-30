variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_osquery_pack" "test" {
  name        = "tf-acc-osquery-pack-404-${var.suffix}"
  description = "Delete idempotency acceptance test pack"

  queries = {
    simple = {
      query = "SELECT 1;"
    }
  }
}
