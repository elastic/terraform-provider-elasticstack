variable "name" {
  type = string
}

variable "owner" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_case_field_definition" "test" {
  owner      = var.owner
  definition = <<-EOT
name: ${var.name}
type: keyword
label: Test Field
control: INPUT_TEXT
EOT
}
