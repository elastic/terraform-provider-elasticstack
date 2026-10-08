variable "list_id" {
  description = "The exception list ID"
  type        = string
}

variable "item_id" {
  description = "The exception item ID"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_exception_list" "test" {
  list_id        = var.list_id
  name           = "Test list for append-only comment semantics"
  description    = "Validates UseStateForUnknown and RequiresReplace plan modifiers."
  type           = "detection"
  namespace_type = "single"
}

resource "elasticstack_kibana_security_exception_item" "test" {
  list_id        = elasticstack_kibana_security_exception_list.test.list_id
  item_id        = var.item_id
  name           = "Append-only comments test item"
  description    = "Used by TestAccResourceExceptionItem_CommentsAppendOnly."
  type           = "simple"
  namespace_type = "single"

  entries = [{
    type     = "match"
    field    = "process.name"
    operator = "included"
    value    = "test-process"
  }]

  # New comment prepended ahead of the existing one — Kibana's PUT
  # returns 400 ("item \"comments\" are append only") for this, so the
  # provider must mark this for replacement.
  comments = [
    { comment = "newly prepended" },
    { comment = "first comment EDITED" },
  ]
}
