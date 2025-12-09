variable "space_id" {
  type = string
}

variable "list_id" {
  type = string
}

variable "name" {
  type = string
}

variable "description" {
  type = string
}

variable "type" {
  type = string
}

variable "meta" {
  type = string
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = "Test Space for Security List"
}

resource "elasticstack_kibana_security_list_data_streams" "test" {
  space_id = elasticstack_kibana_space.test.space_id
}

resource "elasticstack_kibana_security_list" "test" {
  space_id    = elasticstack_kibana_space.test.space_id
  list_id     = var.list_id
  name        = var.name
  description = var.description
  type        = var.type
  meta        = var.meta

  depends_on = [elasticstack_kibana_security_list_data_streams.test]
}
