provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_synthetics_parameter" "test" {
  key                 = "test-key-shared"
  value               = "test-value-shared"
  share_across_spaces = false
}
