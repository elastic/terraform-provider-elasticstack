variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_case_field_definition" "test" {
  owner      = "cases"
  definition = <<-EOT
name: ${var.name}
type: integer
label: Test Field
control: INPUT_NUMBER
EOT
}
