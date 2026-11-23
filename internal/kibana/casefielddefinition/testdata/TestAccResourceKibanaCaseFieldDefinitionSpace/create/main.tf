variable "name" {
  type = string
}

variable "space_id" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test" {
  space_id = var.space_id
  name     = "Case field definition test space"
}

resource "elasticstack_kibana_case_field_definition" "test" {
  space_id   = elasticstack_kibana_space.test.space_id
  owner      = "cases"
  definition = <<-EOT
name: ${var.name}
type: keyword
label: Test Field
control: INPUT_TEXT
EOT
}
