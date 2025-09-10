provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_prebuilt_rule" "enable" {
  tags = ["OS: Linux", "OS: Windows"]
}

resource "elasticstack_kibana_prebuilt_rule" "install_no_enable" {
  tags = []
}
