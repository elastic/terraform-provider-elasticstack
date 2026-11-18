variable "name" {
  type = string
}

variable "other_name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_case_field_definition" "test" {
  owner      = "cases"
  name       = var.name
  definition = <<-EOT
name: ${var.other_name}
type: keyword
label: Test Field
control: INPUT_TEXT
EOT
}
