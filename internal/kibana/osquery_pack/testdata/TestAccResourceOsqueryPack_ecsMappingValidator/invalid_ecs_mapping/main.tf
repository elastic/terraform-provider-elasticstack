provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_osquery_pack" "test" {
  name = "tf-acc-invalid-ecs-mapping"

  queries = {
    invalid = {
      query = "SELECT 1;"
      ecs_mapping = {
        "process.name" = {
          field = "name"
          value = "literal"
        }
      }
    }
  }
}
