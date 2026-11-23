variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_case_field_definition" "test" {
  owner      = "cases"
  name       = var.name
  is_global  = false
  definition = <<-EOT
name: ${var.name}
type: keyword
label: Renamed Label
control: INPUT_TEXT
EOT
}
