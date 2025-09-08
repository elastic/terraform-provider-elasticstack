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
			// Test Create and Read
			{
				Config: testAccResourceKibanaSecurityDetectionRuleCreate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "name", "Test Detection Rule"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "description", "Test security detection rule"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "type", "query"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "severity", "medium"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_detection_rule.test", "id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_detection_rule.test", "rule_id"),
				),
			},
			// Test Update
			{
				Config: testAccResourceKibanaSecurityDetectionRuleUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "name", "Updated Test Detection Rule"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "description", "Updated test security detection rule"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "type", "query"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "severity", "high"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "enabled", "false"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_detection_rule.test", "id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_detection_rule.test", "rule_id"),
				),
			},
			// Test Import (Read)
			{
				ResourceName:      "elasticstack_kibana_security_detection_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
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
  risk        = 50
  enabled     = true
  tags        = ["test"]
  interval    = "5m"
  from        = "now-6m"
  to          = "now"
  version     = 1
  max_signals = 100
}`
}

func testAccResourceKibanaSecurityDetectionRuleUpdate() string {
	return `
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = "Updated Test Detection Rule"
  description = "Updated test security detection rule"
  type        = "query"
  query       = "event.category:network"
  language    = "kuery"
  severity    = "high"
  risk        = 75
  enabled     = false
  tags        = ["test", "updated"]
  interval    = "10m"
  from        = "now-15m"
  to          = "now"
  version     = 1
  max_signals = 200
}`
}
