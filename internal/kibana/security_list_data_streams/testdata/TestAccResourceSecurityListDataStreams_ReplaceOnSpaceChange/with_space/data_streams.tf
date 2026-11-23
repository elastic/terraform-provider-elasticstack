variable "space_id" {
  description = "The space ID"
  type        = string
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = "Test Space for List Data Streams"
}

resource "elasticstack_kibana_security_list_data_streams" "test" {
  space_id = elasticstack_kibana_space.test.space_id
}
