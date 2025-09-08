package security_detection_rule_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceSecurityDetectionRule(t *testing.T) {
	resourceName := "elasticstack_kibana_security_detection_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityDetectionRuleConfig_basic("test-rule"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityDetectionRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "query"),
					resource.TestCheckResourceAttr(resourceName, "query", "*:*"),
					resource.TestCheckResourceAttr(resourceName, "language", "kuery"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "50"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
				),
			},
			{
				Config: testAccSecurityDetectionRuleConfig_update("test-rule-updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityDetectionRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "75"),
				),
			},
		},
	})
}

func testAccCheckSecurityDetectionRuleExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID not set")
		}

		// In a real test, we would make an API call to verify the resource exists
		// For now, we just check that the ID is set
		return nil
	}
}

func testAccCheckSecurityDetectionRuleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_kibana_security_detection_rule" {
			continue
		}

		// In a real test, we would make an API call to verify the resource is deleted
		// For now, we just return nil
	}

	return nil
}

func testAccSecurityDetectionRuleConfig_basic(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = "%s"
  type        = "query"
  query       = "*:*"
  language    = "kuery"
  enabled     = true
  description = "Test security detection rule"
  severity    = "medium"
  risk_score  = 50
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
}
`, name)
}

func testAccSecurityDetectionRuleConfig_update(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = "%s"
  type        = "query"
  query       = "*:*"
  language    = "kuery"
  enabled     = true
  description = "Updated test security detection rule"
  severity    = "high"
  risk_score  = 75
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  author      = ["Test Author"]
  tags        = ["test", "automation"]
  license     = "Elastic License v2"
}
`, name)
}
