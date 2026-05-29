variable "list_id" {}

resource "elasticstack_kibana_security_list_data_streams" "test" {}

resource "elasticstack_kibana_security_list" "test" {
  list_id     = var.list_id
  name        = "Test List for Default Space Items"
  description = "A test security list for items in the default space"
  type        = "keyword"

  depends_on = [elasticstack_kibana_security_list_data_streams.test]
}

resource "elasticstack_kibana_security_list_item" "test" {
  list_id = elasticstack_kibana_security_list.test.list_id
  value   = "default-space-test-value"

  depends_on = [elasticstack_kibana_security_list_data_streams.test]
}
