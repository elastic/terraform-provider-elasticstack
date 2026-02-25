provider "elasticstack" {
  kibana {}
}

# Enable all Windows rules
resource "elasticstack_kibana_security_enable_rule" "windows" {
  space_id = "default"
  key      = "OS"
  value    = "Windows"
}

# Enable rules but don't disable them when the resource is destroyed
resource "elasticstack_kibana_security_enable_rule" "macos_persistent" {
  space_id           = "default"
  key                = "OS"
  value              = "macOS"
  disable_on_destroy = false
}

# Enable all Linux rules
resource "elasticstack_kibana_security_enable_rule" "linux" {
  space_id = "default"
  key      = "OS"
  value    = "Linux"
}

# Enable rules in a custom space
resource "elasticstack_kibana_security_enable_rule" "custom_space" {
  space_id = "security"
  key      = "Data Source"
  value    = "Elastic Defend"
}
