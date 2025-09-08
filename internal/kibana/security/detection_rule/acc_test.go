package detection_rule_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceKibanaSecurityDetectionRule(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceKibanaSecurityDetectionRuleCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "name", "Test Detection Rule"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "description", "Test security detection rule"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "type", "query"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "severity", "medium"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "enabled", "true"),
				),
			},
		},
	})
}

func testAccResourceKibanaSecurityDetectionRuleCreate() string {
	return `
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = "Test Detection Rule"
  description = "Test security detection rule"
  type        = "query"
  query       = "*:*"
  language    = "kuery"
  severity    = "medium"
  enabled     = true
  tags        = ["test"]
  interval    = "5m"
  from        = "now-6m"
  to          = "now"
}`
}