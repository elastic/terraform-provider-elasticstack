provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_osquery_pack" "test" {
  name = "tf-acc-invalid-platform"

  queries = {
    invalid = {
      query    = "SELECT 1;"
      platform = ["ios"]
    }
  }
}
