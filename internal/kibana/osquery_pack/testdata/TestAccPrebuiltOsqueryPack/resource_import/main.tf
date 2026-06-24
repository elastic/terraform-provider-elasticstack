variable "prebuilt_pack_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_osquery_pack" "test" {
  name = "tf-acc-prebuilt-import-should-fail"

  queries = {
    q = {
      query = "SELECT 1;"
    }
  }
}
